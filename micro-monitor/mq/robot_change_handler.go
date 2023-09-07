package mq

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/service"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"micro-common1/biz/cache"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	bizMq "micro-common1/biz/mq"
	"micro-common1/log"
	"micro-common1/redis"
	"micro-common1/util"
	"reflect"
	"time"
)

// 发送任务状态变化通知
func jobStatusCheck(data []bizMq.RobotJobStatus) {
	for _, job := range data {
		info := job.JobInfo
		if enum.JsCancel != info.JobState && enum.JsCompleted != info.JobState {
			continue
		}
		robotId := info.RobotID
		newStatus := dto.RobotStatus{
			RobotId:          robotId,
			OfficeId:         job.OfficeID,
			JobId:            info.JobId,
			GroupId:          info.Jobgroup,
			JobType:          info.JobType,
			TargetPositionId: info.EndPositionID,
		}
		oldStatus, err := cache.GetRobotStatus(job.OfficeID, robotId)
		if err == nil {
			newStatus.RobotName = oldStatus.RobotName
			newStatus.RobotModel = oldStatus.RobotModel
			newStatus.OfficeName = oldStatus.OfficeName
			newStatus.BuildingId = oldStatus.BuildingId
			newStatus.BuildingName = oldStatus.BuildingName
			newStatus.Floor = oldStatus.Floor
		}
		// 时间
		newStatus.Time = time.Now()
		newStatus.NetStatus = enum.NsOnline

		var acceptState enum.AcceptStatusEnum
		if enum.JsCancel == info.JobState {
			newStatus.RobotStatus = enum.RsCancel
		} else {
			// 任务完成
			newStatus.RobotStatus = enum.RsFinished
			acceptState = job.AcceptState
		}
		robot := dao.Robot{Id: robotId, OfficeId: job.OfficeID}
		// 机器人状态检查
		RobotStatusCheck(robot, newStatus, acceptState)

		// 机器人状态变化
		RobotStatusChange(robotId, newStatus)
	}
}

// 机器人状态检查 mq发送信息
func RobotStatusCheck(robot dao.Robot, newStatus dto.RobotStatus, acceptState enum.AcceptStatusEnum) {
	// 获取old
	ctx := context.Background()
	var oldRobotStatus bizMq.RobotStatusUpload
	err := redis.HGetJson(ctx, &oldRobotStatus, constant.LogisticsRobotsStatusUpload, robot.Id)
	if err == nil && newStatus.RobotStatus == oldRobotStatus.Status &&
		newStatus.EstopStatus == oldRobotStatus.EstopStatus && newStatus.PauseType == oldRobotStatus.PauseType {
		return
	}
	// 如果旧的状态为空闲中，新的状态为操作中，那么过滤掉操作中的，不再记录
	if err == nil && oldRobotStatus.JobId != newStatus.JobId && enum.RsIdle == oldRobotStatus.Status &&
		(enum.RsLockForHandle == newStatus.RobotStatus || enum.RsIdle == newStatus.RobotStatus) {
		// 如果结束时间为空，那么设置结束时间
		if oldRobotStatus.StatusEndTime == 0 {
			// 状态上报结束时间为当前状态上报时间
			oldRobotStatus.StatusEndTime = newStatus.Time.UnixNano() / 1e6
			// 计算时长
			oldRobotStatus.TimeConsume = utils.GetTimeDif(oldRobotStatus.StatusStartTime, oldRobotStatus.StatusEndTime)
			// mq发送状态变化
			sendMsg(robot, oldRobotStatus, bizMq.RobotStatusUpload{})
		}
		return
	}
	newRobotStatus := setStatusUploadInfo(oldRobotStatus, newStatus)
	if acceptState != 0 {
		newRobotStatus.AcceptState = acceptState
	}
	if !reflect.DeepEqual(oldRobotStatus, model.RobotStatusUpload{}) {
		// 上报时间为上一个状态的结束时间
		oldRobotStatus.StatusEndTime = newStatus.Time.UnixNano() / 1e6
		// 计算时长
		oldRobotStatus.TimeConsume = utils.GetTimeDif(oldRobotStatus.StatusStartTime, oldRobotStatus.StatusEndTime)
	}
	sendMsg(robot, oldRobotStatus, newRobotStatus)
}

// mq发送状态变化
func sendMsg(robot dao.Robot, oldRobotStatus, newRobotStatus bizMq.RobotStatusUpload) {
	if newRobotStatus.Status == 0 {
		_ = redis.HSetJson(context.Background(), constant.LogisticsRobotsStatusUpload, robot.Id, oldRobotStatus)
	} else {
		_ = redis.HSetJson(context.Background(), constant.LogisticsRobotsStatusUpload, robot.Id, newRobotStatus)
	}
	_ = bizMq.RobotStatusStatistics(bizMq.RobotStatusMqVo{
		RobotId:   robot.Id,
		OldStatus: oldRobotStatus,
		NewStatus: newRobotStatus,
		SentTime:  time.Now().UnixNano() / 1e6,
	})
}

// 设置信息
// 机器人状态为运行异常时，那么状态显示异常；
// 取消任务，显示取消；
// 其他状态显示 正常
func setStatusUploadInfo(oldRobotStatus bizMq.RobotStatusUpload, newStatus dto.RobotStatus) bizMq.RobotStatusUpload {
	statusUpload := toRobotStatusUpload(newStatus)
	// 机器人状态
	switch newStatus.RobotStatus {
	case enum.RsFailed:
		statusUpload.ExecState = enum.EsFailed
	case enum.RsCancel:
		statusUpload.ExecState = enum.EsCancel
	default:
		statusUpload.ExecState = enum.EsNormal
	}
	checkRobotJobStatus(&statusUpload, oldRobotStatus, newStatus)
	return statusUpload
}

func checkRobotJobStatus(statusUpload *bizMq.RobotStatusUpload, oldRobotStatus bizMq.RobotStatusUpload, newStatus dto.RobotStatus) {
	officeId := newStatus.OfficeId
	robotId := newStatus.RobotId
	robotJobId := newStatus.JobId
	var finalJobId string
	var dispatchMode = 0
	var finalJobType = 0
	if reflect.DeepEqual(oldRobotStatus, model.RobotStatusUpload{}) {
		if oldRobotStatus.FinalJobId != "" && oldRobotStatus.JobId == newStatus.JobId {
			finalJobId = oldRobotStatus.FinalJobId
			dispatchMode = oldRobotStatus.DispatchMode
			finalJobType = oldRobotStatus.FinalJobType
		}
	}
	if finalJobId == "" {
		if service.IsAllDispatchMode(officeId) {
			dispatchMode = 1
			// 调度模式，任务会被拆分，获取最后一次发送的任务
			ctx := context.Background()
			robotJobsKey := fmt.Sprintf("%s:%s:%s", constant.DispatchRobotJobJobs, officeId, robotId)
			sendJobDataKey := fmt.Sprintf("%s:%s", constant.CurrentSentJobKey, robotId)
			var jobData model.JobData
			err := redis.HGetJson(ctx, &jobData, robotJobsKey, sendJobDataKey)
			if err == nil {
				log.Infof("robotStatusCheck set dispatch info : %+v ", jobData)
				lastSentJob := jobData.Job
				// 如果任务id，说明机器人执行了调度最后发送的任务
				if lastSentJob.JobId == robotJobId {
					finalJobId = lastSentJob.FinalJobId
					if finalJobId == robotJobId {
						// 并且最终任务的id匹配，说明是最终目的地的任务了
						finalJobType = 1
					}
				}
			}
		}
	}
	log.Infof("dispatch finalJobId :%s;dispatchMode :%d;finalJobType :%d", finalJobId, dispatchMode, finalJobType)
	statusUpload.FinalJobId = finalJobId
	statusUpload.DispatchMode = dispatchMode
	statusUpload.FinalJobType = finalJobType
}

func toRobotStatusUpload(status dto.RobotStatus) bizMq.RobotStatusUpload {
	return bizMq.RobotStatusUpload{
		DocumentId:      util.CreateUUID(),
		RobotId:         status.RobotId,
		OfficeId:        status.OfficeId,
		RobotModel:      string(status.RobotModel),
		Status:          status.RobotStatus,
		JobId:           status.JobId,
		GroupId:         status.GroupId,
		LastUploadTime:  status.Time.UnixNano() / 1e6,
		X:               status.X,
		Y:               status.Y,
		Orientation:     status.Orientation,
		SpotId:          status.LastPositionId,
		Target:          status.TargetPositionId,
		Process:         status.Process,
		NextSpot:        status.NextPositionId,
		Floor:           status.Floor,
		Electric:        status.Electric,
		NetStatus:       status.NetStatus,
		PauseType:       status.PauseType,
		EstopStatus:     status.EstopStatus,
		BuildingId:      status.BuildingId,
		StatusStartTime: status.Time.UnixNano() / 1e6,
	}
}

// 机器人相关状态变化通知
func RobotStatusChange(robotId string, newStatus dto.RobotStatus) {
	// 机器人位置通知
	robotPositionData(newStatus)
	// 机器人信息变化
	robotInfoData(robotId, newStatus)
}

// 机器人信息变化通知
func robotInfoData(robotId string, newStatus dto.RobotStatus) {
	ctx := context.Background()
	var oldStatus dto.RobotStatus
	err := redis.HGetJson(ctx, &oldStatus, constant.LogisticsRobotsInfo, newStatus.RobotId)
	if err != nil && !redis.IsRedisNil(err) {
		return
	}
	var oldStatusStr string
	if !redis.IsRedisNil(err) {
		oldStatusStr = oldStatus.RobotStatus.Description()
	}
	// 机器人信息变化：状态、急停、暂停、网络，位置，电量变化发送mq通知
	if redis.IsRedisNil(err) || newStatus.RobotStatus != oldStatus.RobotStatus ||
		newStatus.EstopStatus != oldStatus.EstopStatus || oldStatus.PauseType != newStatus.PauseType ||
		// 位置（当前位置、目标位置）
		newStatus.LastPositionId != oldStatus.LastPositionId ||
		newStatus.TargetPositionId != oldStatus.TargetPositionId ||
		// 网络状态、电量
		newStatus.NetStatus != oldStatus.NetStatus || newStatus.Electric != oldStatus.Electric {

		log.Infof("发送状态信息变化：robot<%s> 状态(%s -> %s) 急停(%d -> %d) 暂停(%d -> %d) 最终(%s -> %s)"+
			" 目标位置(%s -> %s) 网络(%s -> %s) 电量(%.2f -> %.2f)",
			robotId, oldStatusStr, newStatus.RobotStatus.Description(), oldStatus.EstopStatus, newStatus.EstopStatus,
			oldStatus.PauseType, newStatus.PauseType, oldStatus.LastPositionName, newStatus.LastPositionName,
			oldStatus.TargetPositionName, newStatus.TargetPositionName, oldStatus.NetStatus, newStatus.NetStatus,
			oldStatus.Electric, newStatus.Electric)

		err = bizMq.RobotStatusUpdate(bizMq.RobotStatusChangeMqVo{
			OldStatus: oldStatus,
			NewStatus: newStatus,
			SentTime:  time.Now().UnixNano() / 1e6,
		})
		if err != nil {
			log.WithError(err).Error("发送状态信息变化通知失败")
		}
		// 保存新的状态信息
		_ = redis.HSetJson(ctx, constant.LogisticsRobotsInfo, newStatus.RobotId, newStatus)
	}
}

// 位置变化通知
func robotPositionData(newStatus dto.RobotStatus) {
	ctx := context.Background()
	if newStatus.X == 0 || newStatus.Y == 0 {
		return
	}
	var oldStatus dto.RobotStatus
	err := redis.HGetJson(ctx, &oldStatus, constant.LogisticsRobotsPositionUpload, newStatus.RobotId)
	// 位置变化通知
	if redis.IsRedisNil(err) || newStatus.X != oldStatus.X || newStatus.Y != oldStatus.Y {
		err = bizMq.RobotPositionUpdate(bizMq.RobotStatusChangeMqVo{
			OldStatus: oldStatus,
			NewStatus: newStatus,
			SentTime:  time.Now().UnixNano() / 1e6,
		})
		if err != nil {
			log.WithError(err).Error("发送位置变化通知失败")
		}
		socketData, _ := json.Marshal(newStatus)
		redis.Publish(ctx, constant.WebsocketQueues[4], socketData)
		// 保存新的状态信息
		_ = redis.HSetJson(ctx, constant.LogisticsRobotsPositionUpload, newStatus.RobotId, newStatus)
	}
}
