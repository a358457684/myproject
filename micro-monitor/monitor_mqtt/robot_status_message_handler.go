package monitor_mqtt

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/mq"
	"epshealth-airobot-monitor/service"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"micro-common1/biz/cache"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	"micro-common1/biz/handler"
	"micro-common1/log"
	"micro-common1/redis"
	"micro-common1/util"
	"strconv"
	"strings"
	"time"
)

func robotStatusMessageHandler(topic string, data interface{}) {
	msg := data.(*MqttMsgVo)
	if ok := utils.GetLock(fmt.Sprintf("%s%d", constant.MonitorStatusLock, msg.MsgID), time.Second*3); !ok {
		return
	}

	topics := strings.Split(topic, "/")
	cmdEnum := enum.RobotCmdEnum(topics[5])

	// 消息回复不校验token
	if cmdEnum == enum.RCFeedBack {
		feedbackHandler(msg)
		return
	}

	_, err := utils.ParseToken(msg.Token)
	if err != nil {
		log.WithError(err).Errorf("topic:%s, token:%s校验失败", topic, msg.Token)
		publishInvalidToken(topics, msg.MsgID)
		return
	}

	switch cmdEnum {
	case enum.RCRobotStatus:
		jsonData, _ := json.Marshal(msg.Data)
		var status RobotStatusMessage
		err := json.Unmarshal(jsonData, &status)
		if err != nil {
			log.WithError(err).Errorf("topic:%s, mqtt消息解析失败", topic)
		}
		handleRobotStatus(status, topics, data)
	}
}

// 处理机器人状态信息
func handleRobotStatus(vo RobotStatusMessage, topics []string, data interface{}) {
	robotId := topics[4]
	log.Infof("robot<%s> upload message: %+v", robotId, vo)
	robotStatusProcess(robotId, vo, data)
}

// 真正处理机器人状态的过程
func robotStatusProcess(robotId string, vo RobotStatusMessage, sourceData interface{}) {
	robot := dao.GetByRobotId(robotId)
	if robot.Id == "" {
		log.Errorf("未知机器人:%s", robotId)
		return
	}
	redisKey := constant.LogisticsRobotOnline + robot.Id
	ctx := context.Background()
	exists, err := redis.Exists(ctx, redisKey).Result()
	if err != nil || exists == 0 {
		redis.Set(ctx, redisKey, time.Now().UnixNano(), 0)
	}

	oldStatus, oldStatusErr := cache.GetRobotStatus(robot.OfficeId, robot.Id)

	newStatus := dto.RobotStatus{
		OfficeId:         robot.OfficeId,
		OfficeName:       robot.OfficeName,
		BuildingId:       service.GetBuildingId(robot.OfficeId, vo.Halt),
		BuildingName:     vo.Halt,
		Floor:            vo.Floor,
		RobotId:          robot.Id,
		RobotName:        robot.Name,
		RobotModel:       robot.Model,
		RobotStatus:      vo.Status,
		Electric:         vo.Electric,
		Pause:            vo.PauseType == 1,
		EStop:            vo.EstopStatus == 1,
		LastPositionId:   vo.SpotId,
		LastPositionName: vo.SpotName,
		// NextPositionId:     vo.NextSpotId,
		// NextPositionName:   vo.NextSpotName,
		TargetPositionId:   vo.Target,
		TargetPositionName: vo.TargetName,
		GroupId:            vo.GroupId,
		JobId:              vo.JobId,
		JobType:            vo.JobType,
		Time:               time.Now(),
		NetStatus:          enum.NsOnline,
		X:                  vo.Position.X,
		Y:                  vo.Position.Y,
		Orientation:        vo.Position.Orientation,
		EstopStatus:        vo.EstopStatus, // 老版
		PauseType:          vo.PauseType,   // 老版
	}
	if oldStatusErr == nil {
		newStatus.StartIdleTime = oldStatus.StartIdleTime
		newStatus.StartErrorTime = oldStatus.StartErrorTime
	}

	// 存储当前状态,以及时间,用作状态监控
	setMonitorStatus(robot, vo)
	// 存储当前任务,以及时间,用作任务监控
	setJobMonitorConfig(robot, vo)
	// 存储当前坐标，以及时间,用作范围监控
	setScopeMonitorConfig(robot, vo)
	// 存储当前网络状态，以及时间,用作离线状态监控
	setMonitorNetConnect(robot.OfficeId, robot.Id, enum.NsOnline)
	// 存储当前机器人上传电量，以及时间,用作低电量监控
	setByBatteryMonitorConfig(robot, vo)
	// 检测 工作配置推送是否推送成功 推送 且 没有记录成功 则效验目的地是否是 推送的目的地 如果是 则视为推送成功
	setRobotWorkTimeConfigInfo(robot, vo)
	// 机器人状态检查
	mq.RobotStatusCheck(robot, newStatus, 0)
	// 地图区域处理
	positionUploadHandle(robot, newStatus)

	// 状态或急停状态发生变化时，如果机器人是Y2R或者E2R，将机器人状态发送到pad
	if oldStatusErr != nil || newStatus.RobotStatus != oldStatus.RobotStatus || newStatus.EstopStatus != oldStatus.EstopStatus {
		sendRobotStatusToMobileTerminal(newStatus)
	}

	// 消毒机器人（紫外线或者喷雾），如果任务状态
	if enum.RsFailed == newStatus.RobotStatus {
		for _, robotModel := range constant.DisinfectRobots {
			if robotModel == string(robot.Model) {
				// 修改消毒记录里面的任务状态
				dao.UpdateDisinfectTaskLogEndTime(newStatus.JobId, enum.JsSystemSet)
				break
			}
		}
	}

	// 操作,保存到数据库或缓存中
	if redis.IsRedisNil(oldStatusErr) || newStatus.RobotStatus != oldStatus.RobotStatus {
		log.Infof("状态变化, robotId:%s, %d -> %d", robot.Id, oldStatus.RobotStatus, newStatus.RobotStatus)
		// 如果状态发生变化，则记录到数据库
		doUpdateJob(robot, newStatus)
		// 如果是空闲状态，则记录空闲时间，否则清空空闲时间记录，这里会结束任务
		if newStatus.RobotStatus.IsIdle() && newStatus.StartIdleTime.IsZero() {
			newStatus.StartIdleTime = time.Now()
			log.Infof("notEndToEnd, robotId: %s, status: %s", robot.Id, enum.JsSystemSet.Message())
			// dao.NotEndToEndEx(robot.Id, enum.JsSystemSet)
		}
		if !newStatus.RobotStatus.IsIdle() && !newStatus.RobotStatus.IsLiftStatus() {
			newStatus.StartIdleTime = time.Time{}
		}

		// 如果是任务失败，记录失败开始时间
		if newStatus.RobotStatus == enum.RsFailed {
			if newStatus.StartErrorTime.IsZero() {
				newStatus.StartErrorTime = time.Now()
			}
		} else {
			// 清除失败时间
			newStatus.StartErrorTime = time.Time{}
		}
	} else if newStatus.LastPositionId != "" && newStatus.LastPositionId != oldStatus.LastPositionId {
		// 位置发生变化，记录log
		doAddLog(robot, newStatus)
	}

	// 保存机器人工作状态
	_ = cache.SaveRobotStatus(robot.OfficeId, robot.Id, newStatus)

	doCheckRobotStatusChanged(oldStatus, newStatus, sourceData)

	// 状态调度 机器人状态更新调度处理
	// dispatchRobotStatus(newStatus)

	// 机器人状态变化通过mq通知医护端，医护端通过socket推数据到前端
	mq.RobotStatusChange(robot.Id, newStatus)
}

// 机器人状态变化检测
func doCheckRobotStatusChanged(oldStatus, newStatus dto.RobotStatus, sourceData interface{}) {
	publishRobotToPad(newStatus)
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.LastRobotStatusName, newStatus.OfficeId)
	res, err := redis.HGet(ctx, redisKey, newStatus.RobotId).Result()
	if service.IsStatusChanged(res, err, newStatus) {
		// 存入ES
		service.AddRobotStatusDocument(newStatus, sourceData)
	}
	if service.IsRobotChanged(oldStatus, newStatus) {
		socketData, _ := json.Marshal(newStatus)
		redis.Publish(ctx, constant.WebsocketQueues[0], socketData)
	}
}

func doUpdateJob(robot dao.Robot, s dto.RobotStatus) {
	// 更新job记录，增加jobLog记录
	if s.JobId != "" {
		job := model.RobotJob{
			Id:       s.JobId,
			RobotId:  robot.Id,
			OfficeId: robot.OfficeId,
			JobType:  s.JobType,
			Process:  strings.Join(s.Process, ","),
			GroupId:  s.GroupId,
		}
		old := dao.GetRobotJobById(s.JobId)
		if old.Id == "" {
			job.StartPosition = getStartPosition(s)
			job.EndPosition = s.TargetPositionId
			if job.JobType == 0 {
				// 根据目标位置类型，设置默认任务类型
				job.JobType = getDefaultJobType(s.TargetPositionId, robot.OfficeId)
			}
		} else {
			if old.StartPosition == "" {
				job.StartPosition = getStartPosition(s)
			}
			if old.EndPosition == "" {
				job.EndPosition = s.TargetPositionId
			}
			if old.JobType == 0 && job.JobType == 0 {
				// 根据目标位置类型，设置默认任务类型
				job.JobType = getDefaultJobType(s.TargetPositionId, robot.OfficeId)
			}
		}
		// job.Status = enum.JobStatusEnum(s.RobotStatus.Code())
		job.UpdateDate = time.Now()
		dao.SaveRobotJob(job)
		if s.RobotStatus.NearLift() {
			log.Infof("机器人在电梯旁边，准备发送电梯状态")
			r := model.RobotWithLifts{
				RobotId:  robot.Id,
				OfficeId: robot.OfficeId,
			}
			var liftIds []string
			if robot.OfficeId != "" {
				ids := dao.FindLiftIdsByOfficeId(robot.OfficeId)
				if len(ids) > 0 {
					for _, id := range ids {
						liftIds = append(liftIds, id)
					}
				}
			}
			r.LiftIds = liftIds
			redis.SAdd(context.Background(), constant.NeedLiftStatusRobots, robot.Id)
		} else {
			redis.SRem(context.Background(), constant.NeedLiftStatusRobots, robot.Id)
		}
	}
	doAddLog(robot, s)
}

func doAddLog(robot dao.Robot, s dto.RobotStatus) {
	if s.JobId == "" {
		return
	}
	dao.SaveRobotJobLog(model.RobotJobLog{
		Id:         util.CreateUUID(),
		RobotId:    robot.Id,
		LogType:    s.JobType.Code(),
		Status:     s.RobotStatus.Code(),
		PositionId: s.LastPositionId,
		JobId:      s.JobId,
		CreateDate: time.Now(),
	})
}

// 地图区域处理
func positionUploadHandle(robot dao.Robot, newStatus dto.RobotStatus) {
	officeId := newStatus.OfficeId
	buildingId := newStatus.BuildingId
	floor := newStatus.Floor
	jobId := newStatus.JobId
	startPosition := newStatus.LastPositionId
	endPosition := newStatus.TargetPositionId

	// 1、非空数据判断  机构、楼宇、楼层、任务Id、起始点、目的地、x、y
	if officeId == "" || buildingId == "" || floor == 0 || startPosition == "" || endPosition == "" ||
		jobId == "" || newStatus.X == 0 || newStatus.Y == 0 {
		return
	}

	// 2.非闲时或充电状态才做任务区域记录的处理
	if !newStatus.RobotStatus.IsBusy() {
		return
	}
	endPosition, finalJobId := getFinalEndSpotId(robot.OfficeId, newStatus)

	log.Infof("startPosition:%s; endPosition:%s", startPosition, endPosition)

	// 3、任务起始点或者目的地有一个为空或者起始点与目的地不一致，那么过滤掉
	if startPosition == endPosition {
		return
	}

	log.Infof(`positionUploadHandle info robot<%s>; officeId=%s; buildingId=%s; floor=%d; jobId=%s; startPosition=%s; endPosition=%s`,
		robot.Name, officeId, buildingId, floor, jobId, startPosition, endPosition)

	// 4、获取当前楼层已设置的区域列表 是否已设置区域信息
	mapAreaList := dao.FindMapAreaList(officeId, buildingId, floor)
	if len(mapAreaList) == 0 {
		return
	}

	// 5.根据机构，起始点和目的地查询是否存在
	robotMapArea := dao.FindMapAreaListByOfficeIdAndPosition(officeId, startPosition, endPosition)
	if (robotMapArea != model.RobotMapArea{}) && finalJobId != robotMapArea.JobId {
		return
	}

	for _, area := range mapAreaList {
		if isInThenArea(newStatus, area) {
			var robotJobAreaId string
			if (robotMapArea != model.RobotMapArea{}) {
				robotJobAreaId = robotMapArea.Id
			}
			log.Infof("jobId:%s; areaId:%s; robotJobAreaId:%s", jobId, area.Id, robotJobAreaId)
			if (robotMapArea == model.RobotMapArea{}) {
				service.SaveRobotJobArea(newStatus, finalJobId, area.Id, endPosition)
				return
			}
			if jobId == robotMapArea.JobId {
				// 是否绑定区域
				jobRelation := dao.FindAreaJobRelationByAreaIdAndJobId(area.Id, jobId)
				var jobRelationId string
				if jobRelation.Id != "" {
					jobRelationId = jobRelation.Id
				}
				log.Infof("jobId:%s ;areaId:%s; jobRelationId:%s", jobId, area.Id, jobRelationId)
				if jobRelation.Id == "" {
					service.SaveAreaJobRelation(newStatus, finalJobId, area.Id)
				}
				// 如果区域Id不同
				if jobRelation.Id != "" && robotMapArea.AreaJobId != jobRelation.Id {
					// 设置出区域时间
					outAreaJobRelation := dao.GetAreaJobRelation(robotMapArea.AreaJobId)
					if outAreaJobRelation.Id != "" && outAreaJobRelation.EndTime.IsZero() {
						outAreaJobRelation.EndTime = time.Now()
						_ = dao.UpdateAreaJobRelationEndTimeById(outAreaJobRelation.Id, outAreaJobRelation.EndTime)
					}
					// 设置区域最后一个关联任务区域Id
					robotMapArea.AreaJobId = jobRelation.Id
					dao.UpdateAreaJobIdById(robotMapArea)
				}
			}
			break
		}
	}
}

// 判断point是否在areaPoints形成的区域之内
func isInThenArea(s dto.RobotStatus, area model.RobotMapArea) bool {
	var points []handler.Point
	_ = json.Unmarshal([]byte(area.AreaCoord), &points)
	isInArea, _ := handler.IsPointInArea(s.BuildingId, s.Floor, s.RobotModel, area.RobotModel,
		handler.Point{X: s.X, Y: s.Y}, points...)
	return isInArea
}

func getFinalEndSpotId(officeId string, s dto.RobotStatus) (string, string) {
	var finalEndSpotId, finalJobId string
	// 调度模式
	if service.IsAllDispatchMode(officeId) {
		redisKey := fmt.Sprintf("%s:%s:%s", constant.DispatchRobotJobJobs, officeId, s.RobotId)
		hashKey := fmt.Sprintf("%s:%s", constant.LastSentJobKey, s.RobotId)
		var jobData model.JobData
		err := redis.HGetJson(context.Background(), &jobData, redisKey, hashKey)
		if err == nil {
			finalEndSpotId = jobData.Job.FinalEndSpotId
			finalJobId = jobData.Job.FinalJobId
		}
	}
	if finalEndSpotId == "" {
		finalEndSpotId = s.TargetPositionId
	}
	if finalJobId == "" {
		finalJobId = s.JobId
	}
	return finalEndSpotId, finalJobId
}

// 缓存状态,监测状态异常
func setMonitorStatus(robot dao.Robot, vo RobotStatusMessage) {
	// 获取 monitor_cache 存储的状态
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.MonitorStatusConfigKey, robot.OfficeId)
	var monitorStatusVo model.MonitorStatusVo
	err := redis.HGetJson(ctx, &monitorStatusVo, redisKey, robot.Id)
	// 没有数据 或 状态发生变化则更新状态
	if redis.IsRedisNil(err) || monitorStatusVo.Status != vo.Status {
		_ = redis.HSetJson(ctx, redisKey, robot.Id, model.MonitorStatusVo{
			Status: vo.Status,
			MonitorBaseInfo: model.MonitorBaseInfo{
				StartTime: time.Now(),
				OfficeId:  robot.OfficeId,
				RobotId:   robot.Id,
			},
		})
	}
}

/**
 * set 存储任务配置、时间 监控用
 * 任务ID 为空的时候 移除 monitor_cache 中的监控配置 然后结束
 * monitor_cache 中没有监控配置 则存储 monitor_cache 中有监控配置 则判断任务ID 是否发生变化
 * 变化了则 存储新数据 没变化 则查看是否有配置任务监控配置
 * 没有配置 则检测状态是否发送变化 发生变化 则存储新数据
 * 有配置 则检测上传状态是否配置为检测状态  为检测状态 则不更新时间 不为检测状态 则需要更新时间
 */
func setJobMonitorConfig(robot dao.Robot, robotStatus RobotStatusMessage) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.MonitorJobConfigKey, robot.OfficeId)
	res, err := redis.HGet(ctx, redisKey, robot.Id).Result()
	if robotStatus.JobId == "" {
		// 如果任务id为空 则删除redis 中的存储
		if redis.IsRedisNil(err) {
			redis.Del(ctx, redisKey, robot.Id)
		}
		return
	}
	jobMonitorVo := model.JobMonitorVo{
		JobId:  robotStatus.JobId,
		Status: robotStatus.Status,
		MonitorBaseInfo: model.MonitorBaseInfo{
			StartTime: time.Now(),
			RobotId:   robot.Id,
			OfficeId:  robot.OfficeId,
		},
	}
	// 1 为急停状态
	if robotStatus.EstopStatus == 1 {
		jobMonitorVo.Status = enum.RsStop
	}
	// 获取 monitor_cache 存储的任务配置
	if err == nil {
		var jobMonitorVoCache model.JobMonitorVo
		_ = json.Unmarshal([]byte(res), &jobMonitorVoCache)
		// 存在配置则判断任务ID是否改变 如果改变 则直接存储新任务
		if jobMonitorVo.JobId == jobMonitorVoCache.JobId {
			// 没有改变任务ID 则获取任务监控配置
			jobConfig := dao.GetByOfficeIdAndRobotIdAndMonitorType(robot.OfficeId, robot.Id, "1")
			if (jobConfig == model.JobScopeMonitorConfig{}) {
				// 没有监控配置 则判断当前状态和存储的状态是否一样  一样不更新时间  不一样更新时间
				if jobMonitorVo.Status == jobMonitorVoCache.Status {
					jobMonitorVo.StartTime = jobMonitorVoCache.StartTime
					jobMonitorVo.PushCount = jobMonitorVoCache.PushCount
				}
			} else {
				statusList := strings.Split(jobConfig.RobotStatus, ",")
				// 任务ID 没变 则判断机器人状态  如果状态为配置的状态 则继续计时
				for _, status := range statusList {
					if status == jobMonitorVo.Status.String() {
						jobMonitorVo.StartTime = jobMonitorVoCache.StartTime
						jobMonitorVo.PushCount = jobMonitorVoCache.PushCount
						break
					}
				}
			}
		}
	}
	// 没有数据 则直接存储
	_ = redis.HSetJson(ctx, redisKey, robot.Id, jobMonitorVo)
}

// set 存储范围配置、时间 监控用
// 没有坐标位置属于异常情况 不监控
// 没有配置范围监控 无法监控
// 获取redis 中的 范围监控数据 没有则直接存储
// 有则 判断当前X , Y 坐标 和 配置范围监控坐标  如果没有范围监控配置 则判断坐标是否变化  变化则更新
func setScopeMonitorConfig(robot dao.Robot, robotStatus RobotStatusMessage) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.MonitorScopeConfigKey, robot.OfficeId)
	res, err := redis.HGet(ctx, redisKey, robot.Id).Result()
	// 没有坐标位置 清除范围检测
	if (robotStatus.Position == Position{}) {
		if redis.IsRedisNil(err) {
			redis.Del(ctx, redisKey, robot.Id)
		}
		return
	}
	jobConfig := dao.GetByOfficeIdAndRobotIdAndMonitorType(robot.OfficeId, robot.Id, "2")
	// 没有配置监控 则清除范围检测
	if (jobConfig == model.JobScopeMonitorConfig{}) {
		if redis.IsRedisNil(err) {
			redis.Del(ctx, redisKey, robot.Id)
		}
		return
	}
	scopeMonitorVo := model.ScopeMonitorVo{
		X:      robotStatus.Position.X,
		Y:      robotStatus.Position.Y,
		Status: robotStatus.Status,
		MonitorBaseInfo: model.MonitorBaseInfo{
			StartTime: time.Now(),
			RobotId:   robot.Id,
			OfficeId:  robot.OfficeId,
		},
	}
	// 1 为急停状态
	if robotStatus.EstopStatus == 1 {
		scopeMonitorVo.Status = enum.RsStop
	}
	if err == nil {
		var scopeMonitorVoCache model.ScopeMonitorVo
		_ = json.Unmarshal([]byte(res), &scopeMonitorVo)
		configX := scopeMonitorVoCache.X
		configY := scopeMonitorVoCache.Y
		currentX := scopeMonitorVo.X
		currentY := scopeMonitorVo.X
		statusList := strings.Split(jobConfig.RobotStatus, ",")
		for _, status := range statusList {
			if status == scopeMonitorVo.Status.String() {
				distance := service.GetPositionDistance(currentX, currentY, configX, configY)
				monitorScope, err := strconv.Atoi(jobConfig.MonitorScope)
				if err != nil {
					monitorScope = 3
				}
				if distance < float64(monitorScope) {
					// 如果坐标半径在配置距离之内 且状态为配置中的状态 则累计时间
					scopeMonitorVo.X = scopeMonitorVoCache.X
					scopeMonitorVo.Y = scopeMonitorVoCache.Y
					scopeMonitorVo.StartTime = scopeMonitorVoCache.StartTime
					scopeMonitorVo.PushCount = scopeMonitorVoCache.PushCount
				}
				break
			}
		}
	}
	// 没有数据 则直接存储
	_ = redis.HSetJson(ctx, redisKey, robot.Id, scopeMonitorVo)
}

// set 存储网络连接状态、时间 监控用
func setMonitorNetConnect(officeId, robotId string, netStatus enum.NetStatusEnum) {
	// 获取 monitor_cache 存储的状态
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.MonitorNetConnectStatusKey, officeId)
	var monitorNetConnectVo model.MonitorNetConnectVo
	err := redis.HGetJson(ctx, &monitorNetConnectVo, redisKey, robotId)
	// 没有数据 或 状态发生变化则更新状态
	if redis.IsRedisNil(err) || monitorNetConnectVo.NetStatus != netStatus {
		_ = redis.HSetJson(ctx, redisKey, robotId, model.MonitorNetConnectVo{
			NetStatus: netStatus,
			MonitorBaseInfo: model.MonitorBaseInfo{
				StartTime: time.Now(),
				OfficeId:  officeId,
				RobotId:   robotId,
			},
		})
	}
}

// set 存储低电量配置、时间 监控用
func setByBatteryMonitorConfig(robot dao.Robot, robotStatus RobotStatusMessage) {
	byBatteryMonitorVo := model.ByBatteryMonitorVo{
		Electric: robotStatus.Electric,
		MonitorBaseInfo: model.MonitorBaseInfo{
			StartTime: time.Now(),
			OfficeId:  robot.OfficeId,
			RobotId:   robot.Id,
		},
	}
	var byBatteryMonitorVoCache model.ByBatteryMonitorVo
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.MonitorElectricConfigKey, robot.OfficeId)
	// 不是充电中继续存储
	if enum.RsCharging != robotStatus.Status {
		err := redis.HGetJson(ctx, &byBatteryMonitorVoCache, redisKey, robot.Id)
		if err == nil {
			byBatteryMonitorVo.PushCount = byBatteryMonitorVoCache.PushCount
			byBatteryMonitorVo.StartTime = byBatteryMonitorVoCache.StartTime
		}
	}
	// 没有数据 或 电量信息发生变化则更新状态
	if (byBatteryMonitorVoCache == model.ByBatteryMonitorVo{}) ||
		byBatteryMonitorVo.Electric != byBatteryMonitorVoCache.Electric ||
		byBatteryMonitorVo.PushCount != byBatteryMonitorVoCache.PushCount {

		_ = redis.HSetJson(ctx, redisKey, robot.Id, byBatteryMonitorVo)
	}
}

// 检测 工作配置推送是否推送成功 推送 且 没有记录成功 则效验目的地是否是 推送的目的地 如果是 则视为推送成功
func setRobotWorkTimeConfigInfo(robot dao.Robot, robotStatus RobotStatusMessage) {
	// 获取 monitor_cache 存储的状态
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", constant.RobotWorkPushStatusInfo, robot.OfficeId)
	var workPushInfo model.RobotWorkTimeConfigInfo
	err := redis.HGetJson(ctx, &workPushInfo, redisKey, robot.Id)
	if err == nil {
		if workPushInfo.IsPush && !workPushInfo.IsPushSuccess && robotStatus.Target == workPushInfo.PushPositionGuId {
			workPushInfo.IsPushSuccess = true
			workPushInfo.PushSuccessTime = time.Now().UnixNano()
			_ = redis.HSetJson(ctx, redisKey, robot.Id, workPushInfo)
			log.Infof("工作任务已经推送成功,机构<%s> 机器人<%s> 推送信息为<%+v>", robot.OfficeId, robot.Id, workPushInfo)
		}
	}
}

// 获取起始点id
func getStartPosition(s dto.RobotStatus) string {
	if len(s.Process) > 0 {
		return s.Process[0]
	}
	return s.LastPositionId
}

// 根据目标位置类型，获取默认任务类型，防止底盘上传位置异常出现 NullPointerException
func getDefaultJobType(officeId, positionGuid string) enum.JobTypeEnum {
	p := dao.GetRobotPositionByGuId(positionGuid, officeId)
	if p.Id != "" {
		if enum.PtCharge == p.PositionType {
			return enum.JtCharge
		}
		if enum.PtInit == p.PositionType {
			return enum.JtBack
		}
	}
	return enum.JtDistribute
}

// 给PAD端发送状态
func sendRobotStatusToMobileTerminal(robotStatus dto.RobotStatus) {
	for _, robotModel := range constant.PadRobots {
		if robotModel == string(robotStatus.RobotModel) {
			officeId := robotStatus.OfficeId
			if officeId == "" {
				log.Infof("officeId为空，不通知到pad，robotId:%s", robotStatus.RobotId)
				return
			}
			mobileTerminalList := dao.FindMobileTerminalByOfficeId(officeId)
			if len(mobileTerminalList) == 0 {
				return
			}
			robotStatusList, err := cache.FindRobotStatusByOfficeId(officeId)
			if err != nil {
				return
			}
			var robotStatusShowVos []dto.RobotStatusToPAD
			for _, status := range robotStatusList {
				if status.RobotId == robotStatus.RobotId {
					// 最新的状态
					robotStatusShowVos = append(robotStatusShowVos, dto.NewRobotStatusToPAD(robotStatus))
					continue
				}
				robotStatusShowVos = append(robotStatusShowVos, dto.NewRobotStatusToPAD(status))
			}
			publishPadToOffice(officeId, 0, robotStatusShowVos)
			log.Infof("mqttSender send RobotStatus message topic : /pad/toOffice/%s", officeId)
			return
		}
	}
}
