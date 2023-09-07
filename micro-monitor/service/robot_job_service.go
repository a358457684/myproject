package service

import (
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"fmt"
	"micro-common1/biz/cache"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	"micro-common1/biz/manager"
	"micro-common1/log"
	"strconv"
	"time"
)

// 当前发送给机器人执行的任务
type SentRobotJobGroupData struct {
	OfficeId   string             `json:"officeId"`   // 机构id
	RobotId    string             `json:"robotId"`    // 机器人id
	RobotModel manager.RobotType  `json:"robotModel"` // 机器人类型
	Timestamp  int64              `json:"timestamp"`  // 时间戳
	Jobs       []model.RobotJobVo `json:"jobs"`       // 机构id
	EndType    int                `json:"endType"`    // 结束类型
}

type JobVo struct {
	JobId                     string             `json:"jobId"`                     // 任务id
	GroupId                   string             `json:"groupId"`                   // 任务组id
	JobNo                     string             `json:"jobNo"`                     // 任务编码
	JobType                   enum.JobTypeEnum   `json:"jobType"`                   // 任务类型
	JobTypeText               string             `json:"jobTypeText"`               // 任务类型描述
	StartSpotId               string             `json:"startSpotId"`               // 起始位置id
	StartSpotFullName         string             `json:"startSpotFullName"`         // 起始位置全名
	EndSpotId                 string             `json:"endSpotId"`                 // 目标位置id
	EndSpotFullName           string             `json:"endSpotFullName"`           // 目标位置全名
	StartPositionBuildingName string             `json:"startPositionBuildingName"` // 起始位置楼宇
	EndPositionBuildingName   string             `json:"endPositionBuildingName"`   // 起始位置楼宇
	StartUserId               string             `json:"startUserId"`               // 创始人id
	StartUserName             string             `json:"startUserName"`             // 创始人名称
	RobotId                   string             `json:"robotId"`                   // 机器人id
	RobotName                 string             `json:"robotName"`                 // 机器人名称
	JobStatus                 enum.JobStatusEnum `json:"jobStatus"`                 // 任务状态
	JobStatusText             string             `json:"jobStatusText"`             // 任务状态描述
	Origin                    string             `json:"origin"`                    // 任务来源
}

// 获取机器人任务信息
func FindPageRobotJob(vo model.RobotJobQueryVo) model.PageResult {

	jobVos := make([]JobVo, 0, 10)
	robots, buildings, robotUserMap, positionMap := getBasicData(vo.OfficeId)

	if vo.PageIndex == 1 && (vo.EndDate.IsZero() || vo.EndDate.After(time.Now())) {
		// 获取机构下所有正在执行和排队中的任务
		jobList, _ := cache.FindRobotJobByOfficeID(vo.OfficeId)
		for _, job := range jobList {
			if vo.JobType != 0 && job.JobType != vo.JobType {
				continue
			}
			if vo.JobStatus != 0 && job.JobState != vo.JobStatus {
				continue
			}
			jobVo := JobVo{
				JobNo:                     getJobNo(job.JobId, job.CreateTime.Time),
				JobId:                     job.JobId,
				GroupId:                   job.Jobgroup,
				JobType:                   job.JobType,
				JobTypeText:               job.JobType.String(),
				RobotId:                   job.RobotID,
				RobotName:                 getRobotName(robots, job.RobotID),
				StartPositionBuildingName: getBuildingName(buildings, job.StartBuild),
				EndPositionBuildingName:   getBuildingName(buildings, job.EndBuildId),
				StartSpotFullName:         job.StartPosName,
				EndSpotFullName:           job.EndPosName,
				StartUserId:               job.JobCreator,
				StartUserName:             getUsername(robotUserMap, job.JobCreator),
				Origin:                    job.Origin.String(),
				JobStatus:                 job.JobState,
				JobStatusText:             job.JobState.Message(),
			}
			jobVos = append(jobVos, jobVo)
		}
	}

	// 已完成的任务
	finishJob, total := dao.FindRobotJobByOffice(vo)
	robotName := getRobotName(robots, vo.RobotId)
	for _, job := range finishJob {
		jobVo := setRobotJobInfo(job, robotName, buildings, positionMap, robotUserMap)
		jobVos = append(jobVos, jobVo)
	}

	return model.PageResult{
		PageIndex: vo.PageIndex,
		PageSize:  vo.PageSize,
		Total:     total,
		Data:      jobVos,
	}
}

func getUsername(robotUserMap map[string]string, userId string) string {
	username := robotUserMap[userId]
	if username != "" {
		return username
	}
	if userId == constant.UserAdminId {
		return constant.UserSystemName
	}
	if userId != "" {
		return userId
	}
	return constant.AdminName
}

func getBasicData(officeId string) ([]dto.RobotStatus, []dao.OfficeBuildingVo, map[string]string, map[string]dao.RobotPosition) {
	robots, _ := cache.FindRobotStatusByOfficeId(officeId)
	buildings := dao.GetBuildingByOfficeId(officeId)
	robotUsers := dao.FindRobotUserByOfficeId(officeId)
	robotUserMap := make(map[string]string, len(robotUsers))
	for _, robotUser := range robotUsers {
		if robotUser.Nickname == "" {
			robotUserMap[robotUser.Id] = robotUser.Username
			continue
		}
		robotUserMap[robotUser.Id] = robotUser.Nickname
	}
	positions := dao.FindRobotPositionByOfficeId(officeId)
	positionMap := make(map[string]dao.RobotPosition, len(positions))
	for _, position := range positions {
		positionMap[position.GuId] = position
	}
	return robots, buildings, robotUserMap, positionMap
}

// 是否开启了调度模式（所有的调度模式）
func IsAllDispatchMode(officeId string) bool {
	oc := getOfficeConfig(officeId)
	dispatch := isAllDispatchModeByOfficeConfig(oc)
	logDispatchMode(officeId, oc, dispatch)
	return dispatch
}

func logDispatchMode(officeId string, config dao.OfficeConfig, canDispatch bool) {
	var robotId string
	var mode int
	if config.Id != "" {
		if config.RobotId != "" {
			robotId = config.RobotId
		}
		mode = int(config.Mode)
	}
	key := fmt.Sprintf("%s:%s", officeId, robotId)
	if canRetry(key, strconv.FormatBool(canDispatch), constant.OfficeConfigLogTimespan) {
		var msg string
		if canDispatch {
			msg = "使用调度系统"
		} else {
			msg = "不使用调度系统"
		}
		log.Infof("%s, officeId: %s, robotId: %s, mode: %d", msg, officeId, robotId, mode)
	}
}

var robotRecoverLastRetryTimeMap = make(map[string]map[string]int64)

/**
 * 能不能进行重试
 * @param key              key
 * @param cmpKey           比较的key，判断值是否改变了
 * @param waitingRetrySpan 秒，重试时间间隔
 */
func canRetry(key string, cmpKey string, waitingRetrySpan int) bool {
	if lastRetryTime, ok := robotRecoverLastRetryTimeMap[key]; ok {
		millisecond := time.Now().UnixNano() / 1e6
		if len(lastRetryTime) != 0 {
			for key, value := range lastRetryTime {
				if key == cmpKey {
					if isExpiredTime(value, millisecond, waitingRetrySpan) {
						robotRecoverLastRetryTimeMap[key] = map[string]int64{cmpKey: millisecond}
						return true
					}
				} else {
					// 如果cmpKey已经改变了，说明可以进行重试了
					robotRecoverLastRetryTimeMap[key] = map[string]int64{cmpKey: millisecond}
					return true
				}
			}
		} else {
			robotRecoverLastRetryTimeMap[key] = map[string]int64{cmpKey: millisecond}
			return true
		}
	}
	return false
}

// 是否开启了调度模式（所有的调度模式）
func isAllDispatchModeByOfficeConfig(oc dao.OfficeConfig) bool {
	return isDispatchMode(oc, int(enum.DmDispatch))
}

/**
 * 是否为使用调度系统
 */
func isDispatchMode(oc dao.OfficeConfig, dispatchMode int) bool {
	if oc.Id != "" && int(oc.Mode) == dispatchMode {
		return true
	}
	return false
}

// 先不做这种处理，调度模式的时候，必须有一个默认的机构配置，并且配置为调度模式的
// 如果没有获取到机构默认的机构配置，那么获取机器人的
func getOfficeConfig(officeId string) dao.OfficeConfig {
	// 这里先没有使用缓存，可以优化
	officeConfigs := dao.FindOfficeConfigByOffice(officeId)
	if len(officeConfigs) != 0 {
		for _, config := range officeConfigs {
			if config.RobotId == "" {
				return config
			}
		}
	}
	return dao.OfficeConfig{}
}

func setRobotJobInfo(job model.RobotJob, robotName string, buildings []dao.OfficeBuildingVo,
	positionMap map[string]dao.RobotPosition, robotUserMap map[string]string) JobVo {

	return JobVo{
		RobotId:                   job.RobotId,
		JobNo:                     getJobNo(job.Id, job.CreateDate),
		JobId:                     job.Id,
		JobStatus:                 job.Status,
		JobStatusText:             job.Status.Message(),
		RobotName:                 robotName,
		GroupId:                   job.GroupId,
		JobType:                   job.JobType,
		JobTypeText:               job.JobType.String(),
		StartPositionBuildingName: getBuildingName(buildings, positionMap[job.StartPosition].BuildingId),
		EndPositionBuildingName:   getBuildingName(buildings, positionMap[job.EndPosition].BuildingId),
		StartSpotId:               job.StartPosition,
		StartSpotFullName:         positionMap[job.StartPosition].FullName,
		EndSpotId:                 job.EndPosition,
		EndSpotFullName:           positionMap[job.EndPosition].FullName,
		StartUserId:               job.StartUserId,
		StartUserName:             getUsername(robotUserMap, job.StartUserId),
		Origin:                    job.Origin.String(), // 任务来源
	}
}

func isExpiredTime(timestamp int64, now int64, expirySpan int) bool {
	if timestamp != 0 {
		return (now - timestamp) > int64(expirySpan*1000)
	}
	return true
}

// 获取任务编码
func getJobNo(jobId string, dateTime time.Time) string {
	createDateStr := dateTime.Format(constant.DateTimeFormat)
	if len(jobId) > 8 {
		return fmt.Sprintf("%s-%s", createDateStr, jobId[0:8])
	}
	return fmt.Sprintf("%s-%s", createDateStr, jobId)
}
