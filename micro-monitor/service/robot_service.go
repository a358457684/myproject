package service

import (
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"fmt"
	"math"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	"micro-common1/redis"
)

// 默认楼宇
func GetBuildingId(officeId, halt string) string {
	if halt == "" {
		halt = constant.DefaultBuildingHalt
	}
	return fmt.Sprintf("%s__%s", officeId, halt)
}

// 获取两点之间的直线距离
func GetPositionDistance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	x := math.Abs(x1 - x2)
	y := math.Abs(y1 - y2)
	// 开平方
	return math.Sqrt(x*x + y*y)
}

// 机器人状态是否改变
func IsStatusChanged(res string, err error, robot dto.RobotStatus) bool {
	if redis.IsRedisNil(err) {
		return true
	}
	var lastRobot dto.RobotStatus
	err = json.Unmarshal([]byte(res), &lastRobot)
	return err == nil &&
		// 状态
		(robot.RobotStatus != lastRobot.RobotStatus ||
			// 网络状态
			robot.NetStatus != lastRobot.NetStatus ||
			// 任务位置
			robot.LastPositionId != lastRobot.LastPositionId ||
			// 急停状态
			robot.EStop != lastRobot.EStop ||
			// 暂停状态
			robot.Pause != lastRobot.Pause) &&
		// 而且上传时间要大于最后的时间
		robot.Time.After(lastRobot.Time)
}

func IsRobotChanged(oldStatus, newStatus dto.RobotStatus) bool {
	return oldStatus.RobotStatus != newStatus.RobotStatus || oldStatus.NetStatus != newStatus.NetStatus ||
		oldStatus.Electric != newStatus.Electric || oldStatus.EStop != newStatus.EStop ||
		oldStatus.PauseType != newStatus.PauseType || oldStatus.RobotName != newStatus.RobotName ||
		oldStatus.RobotModel != newStatus.RobotModel
}

// 封装返回给前端的机器人状态
func ToWebRobotStatus(robot dao.Robot, configs []dao.OfficeConfig, robotStatus dto.RobotStatus) model.RobotStatusVo {
	statusVo := model.RobotStatusVo{
		RobotId:             robotStatus.RobotId,
		Name:                robot.Name,
		RobotModel:          robotStatus.RobotModel,
		RobotAccount:        robot.Account,
		OfficeId:            robotStatus.OfficeId,
		OfficeName:          robot.OfficeName,
		BuildingId:          robotStatus.BuildingId,
		BuildingName:        robotStatus.BuildingName,
		Floor:               robotStatus.Floor,
		Status:              robotStatus.RobotStatus,
		StatusText:          robotStatus.RobotStatus.Description(),
		Electric:            robotStatus.Electric,
		NetStatus:           robotStatus.NetStatus.Code(),
		NetStatusText:       robotStatus.NetStatus.Description(),
		X:                   robotStatus.X,
		Y:                   robotStatus.Y,
		LastUploadTime:      robotStatus.Time,
		ChassisSerialNumber: robot.ChassisSerialNumber,
		SoftVersion:         robot.SoftVersion,
		DispatchMode:        false,
		PauseType:           robotStatus.PauseType,
		EStopStatus:         robotStatus.EstopStatus,
	}
	for _, config := range configs {
		if config.OfficeId == robot.OfficeId && config.RobotId == "" {
			statusVo.DispatchMode = config.Mode == enum.DmDispatch
			break
		}
	}
	return statusVo
}

// 封装未连接机器人信息
func ToWebOffLineRobotInfo(robot dao.Robot, configs []dao.OfficeConfig, notConnect int) model.RobotStatusVo {
	statusVo := model.RobotStatusVo{
		RobotId:             robot.Id,
		Name:                robot.Name,
		RobotModel:          robot.Model,
		NetStatus:           notConnect,
		NetStatusText:       "未连接",
		X:                   0,
		Y:                   0,
		RobotAccount:        robot.Account,
		ChassisSerialNumber: robot.ChassisSerialNumber,
		SoftVersion:         robot.SoftVersion,
		OfficeId:            robot.OfficeId,
		OfficeName:          robot.OfficeName,
		DispatchMode:        false,
	}
	if statusVo.OfficeId != "" {
		for _, config := range configs {
			if config.OfficeId == robot.OfficeId && config.RobotId == "" {
				statusVo.DispatchMode = config.Mode == enum.DmDispatch
				break
			}
		}
	}
	return statusVo
}

// 获取robot名字
func getRobotName(robots []dto.RobotStatus, robotId string) string {
	for _, robot := range robots {
		if robot.RobotId == robotId {
			return robot.RobotName
		}
	}
	return robotId
}

// 获取楼宇名字
func getBuildingName(buildings []dao.OfficeBuildingVo, buildingId string) string {
	for _, building := range buildings {
		if building.Id == buildingId {
			return building.Name
		}
	}
	return buildingId
}
