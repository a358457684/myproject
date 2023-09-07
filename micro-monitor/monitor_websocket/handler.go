package monitor_websocket

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"fmt"
	baseRedis "github.com/go-redis/redis/v8"
	"micro-common1/biz/cache"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	"micro-common1/biz/handler"
	"micro-common1/biz/manager"
	bizMq "micro-common1/biz/mq"
	"micro-common1/log"
	"micro-common1/redis"
	wsManager "micro-common1/websocket"
	"strconv"
	"time"
)

const (
	// ping相关
	ping        = "ping"
	pongMessage = `{"path":"pong"}`
	// 状态列表菜单
	statusMenu = "status"
	// 任务列表菜单
	jobMenu = "job"
	// 推送列表菜单
	pushMenu = "push"
	// 更新推送列表指定数据
	updatePushMenu = "updatePush"
	// 地图菜单
	monitorMenu = "monitor"
	// 所有机器人
	robotMenu = "robot"
	// 调度资源，隶属地图菜单
	dispatchMenu = "dispatch"
	// 电梯内的坐标 -10000
	liftCoordinate = -10000
)

type ServerData struct {
	Path string      `json:"path"`
	Data interface{} `json:"data"`
}

type RobotStatusRes struct {
	RobotId   string  `json:"robotId"`
	RobotName string  `json:"robotName"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

type RobotRes struct {
	RobotId        string               `json:"robotId"`
	Name           string               `json:"name"`
	Status         enum.RobotStatusEnum `json:"status"`
	StatusText     string               `json:"statusText"`
	EStopStatus    int                  `json:"eStopStatus"`
	PauseType      int                  `json:"pauseType"`
	NetStatus      enum.NetStatusEnum   `json:"netStatus"`
	NetStatusText  string               `json:"netStatusText"`
	Electric       float64              `json:"electric"`
	RobotModel     manager.RobotType    `json:"robotModel"`
	X              float64              `json:"x"`
	Y              float64              `json:"y"`
	BuildingId     string               `json:"buildingId"`
	BuildingName   string               `json:"buildingName"`
	Floor          int                  `json:"floor"`
	LastUploadTime time.Time            `json:"lastUploadTime"`
}

type DispatchDTO struct {
	OfficeId   string                 `json:"officeId"`
	NotifyInfo []bizMq.ResourceNotify `json:"notifyInfo"`
}

// 所有机器人：校验 officeId
func canBroadCastAllRobotData(client wsManager.WsClienter, params ...interface{}) bool {
	wsClient := client.(*RobotWebSocket)
	isOk := wsClient.OfficeId == params[0].(string)
	if isOk {
		log.Infof("实时推送信息%+v", params)
	}
	return isOk
}

// 状态、任务：校验 path、officeId、robotId
func canBroadCastPathData(client wsManager.WsClienter, params ...interface{}) bool {
	wsClient := client.(*RobotWebSocket)
	isOk := wsClient.Path == params[0].(string) &&
		wsClient.OfficeId == params[1].(string) &&
		wsClient.RobotId == params[2].(string)
	if isOk {
		log.Infof("实时推送信息%+v", params)
	}
	return isOk
}

// 位置、调度资源：校验 path、officeId、buildingId、floor
func canBroadCastMonitorData(client wsManager.WsClienter, params ...interface{}) bool {
	wsClient := client.(*RobotWebSocket)
	isOk := wsClient.Path == params[0].(string) &&
		wsClient.OfficeId == params[1].(string) &&
		wsClient.BuildingId == params[2].(string) &&
		wsClient.Floor == params[3].(int)
	if isOk {
		log.Infof("实时推送信息%+v", params)
	}
	return isOk
}

// 位置
func sendPosition(newStatus dto.RobotStatus) {
	// 判断是否在电梯内
	if liftCoordinate == newStatus.X && liftCoordinate == newStatus.Y {
		return
	}
	mapInfo, err := getMapInfo(newStatus)
	if err != nil {
		return
	}
	robotModel := newStatus.RobotModel.Type()
	checkModel := mapInfo.RobotType.Type()
	if robotModel == nil || checkModel == nil || robotModel.Chassis == nil || checkModel.Chassis == nil {
		return
	}
	point := handler.Point{X: newStatus.X, Y: newStatus.Y}
	if robotModel.Chassis.Supplier != checkModel.Chassis.Supplier {
		point, err = handler.ConvertPoint(newStatus.BuildingId, newStatus.Floor, newStatus.RobotModel,
			mapInfo.RobotType, point)
		log.WithError(err).Errorf("多机器地图坐标转换失败")
	}
	point = mapInfo.PixelPointByRealPoint(point, float64(mapInfo.Height))
	_manager.Broadcast(
		ServerData{
			Path: monitorMenu,
			Data: RobotStatusRes{
				RobotId:   newStatus.RobotId,
				RobotName: newStatus.RobotName,
				X:         point.X,
				Y:         point.Y,
			},
		},
		canBroadCastMonitorData,
		monitorMenu, newStatus.OfficeId, newStatus.BuildingId, newStatus.Floor)
}

func getMapInfo(status dto.RobotStatus) (handler.BaseMapInfo, error) {
	redisKey := fmt.Sprintf("%s:%s:%d", constant.MonitorMap, status.BuildingId, status.Floor)
	var mapInfo handler.BaseMapInfo
	err := redis.GetJson(context.Background(), &mapInfo, redisKey)
	return mapInfo, err
}

// 状态
func sendRobotStatus(newStatus model.ElasticRobotStatus) {
	_manager.Broadcast(
		ServerData{
			Path: statusMenu,
			Data: newStatus,
		},
		canBroadCastPathData,
		statusMenu, newStatus.OfficeId, newStatus.RobotId)
}

// 任务，直接刷新
func sendRobotJob(jobs []bizMq.RobotJobStatus) {
	for _, job := range jobs {
		_manager.Broadcast(
			ServerData{
				Path: jobMenu,
				Data: job.JobInfo.RobotID,
			},
			canBroadCastPathData,
			jobMenu, job.OfficeID, job.JobInfo.RobotID)
	}
}

// 推送
func sendRobotPush(isUpdate bool, data model.ElasticRobotPushMessage) {
	path := pushMenu
	if isUpdate {
		path = updatePushMenu
	}
	_manager.Broadcast(
		ServerData{
			Path: path,
			Data: data,
		},
		canBroadCastPathData,
		pushMenu, data.OfficeId, data.RobotId)
}

// 机构全部机器人的基本信息
func sendAllRobotStatus(newStatus dto.RobotStatus) {
	_manager.Broadcast(
		ServerData{
			Path: robotMenu,
			Data: RobotRes{
				RobotId:        newStatus.RobotId,
				Name:           newStatus.RobotName,
				Status:         newStatus.RobotStatus,
				StatusText:     newStatus.RobotStatus.Description(),
				EStopStatus:    newStatus.EstopStatus,
				PauseType:      newStatus.PauseType,
				NetStatus:      newStatus.NetStatus,
				NetStatusText:  newStatus.NetStatus.Description(),
				Electric:       newStatus.Electric,
				RobotModel:     newStatus.RobotModel,
				X:              newStatus.X,
				Y:              newStatus.Y,
				BuildingId:     newStatus.BuildingId,
				BuildingName:   newStatus.BuildingName,
				Floor:          newStatus.Floor,
				LastUploadTime: newStatus.Time,
			},
		},
		canBroadCastAllRobotData,
		newStatus.OfficeId)
}

// 调度资源
func sendDispatchResource(officeId string, notifyInfo []bizMq.ResourceNotify) {
	statusList, err := cache.FindRobotStatusByOfficeId(officeId)
	if err != nil {
		return
	}
	for _, info := range notifyInfo {
		messages := make([]string, 0)
		switch info.ResType {
		case bizMq.ResTypeCtrlArea:
			areaId, maxRobot, robotNames := controlHandler(info, statusList)
			messages = append(messages,
				fmt.Sprintf("[%s]前方经过当前楼宇楼层管制区域[%s]，最大容纳：[%s]，已有机器人[%v]",
					time.Now().Format(constant.DateTimeFormat), areaId, maxRobot, robotNames))

		case bizMq.ResTypeLiftCtrlArea:
			areaId, maxRobot, robotNames := controlHandler(info, statusList)
			messages = append(messages,
				fmt.Sprintf("[%s]前方经过当前楼宇楼层电梯管制区域[%s]，最大容纳：[%s]，已有机器人[%v]",
					time.Now().Format(constant.DateTimeFormat), areaId, maxRobot, robotNames))

		case bizMq.ResTypeLift:
			res := info.LiftResNotify
			deviceSn := dao.FindLiftDeviceSnById(res.LiftID)
			robotName := getRobotStatus(statusList, res.CurrentRobot.RobotID)
			messages = append(messages, fmt.Sprintf("[%s]梯控号[%s]电梯当前[%s]在搭乘，还有[%d]台机器人在等待点排队中",
				time.Now().Format(constant.DateTimeFormat), deviceSn, robotName, len(res.CurrentUseWaitPoints)))

		case bizMq.ResTypeSafeWait:
			res := info.SafeResNotify
			positions := dao.FindRobotPositionByCondition(model.OfficeFloorVo{
				OfficeId:   officeId,
				BuildingId: info.BuildID,
				Floor:      int(info.Floor),
			})
			for _, info := range res.CurrentUseWaitPoints {
				robotName := getRobotStatus(statusList, info.RobotID)
				positionName := getPositionName(positions, info.WaitPointID)
				messages = append(messages, fmt.Sprintf("[%s]%s与机器人相遇超过危险距离，调度移动到[%s]停靠等待",
					time.Now().Format(constant.DateTimeFormat), robotName, positionName))
			}

		case bizMq.ResTypeFloor:
			res := info.FloorResNotify
			robotNames := make([]string, 0)
			for _, robot := range res.CurrentRobots {
				name := getRobotStatus(statusList, robot.RobotID)
				robotNames = append(robotNames, name)
			}
			messages = append(messages, fmt.Sprintf("[%s]%d楼楼层管制，已有机器人[%v]",
				time.Now().Format(constant.DateTimeFormat), info.Floor, robotNames))

		default:
			log.Errorf("未知的调度处理类型:%d", info.ResType)
			continue
		}
		log.Infof("=====推送消息到: menu:%s; 机构:%s; 推送内容: %v", dispatchMenu, officeId, messages)
		_manager.Broadcast(
			ServerData{
				Path: dispatchMenu,
				Data: messages,
			},
			canBroadCastMonitorData,
			monitorMenu, officeId, info.BuildID, info.Floor)
	}
}

// 管制区域和电梯管制区域的处理
func controlHandler(info bizMq.ResourceNotify, statusList []dto.RobotStatus) (string, string, []string) {
	res := info.AreaResNotify
	areaId := res.CtrlAreaID
	robotNames := make([]string, 0)
	for _, robot := range res.CurrentRobotIDs {
		robotNames = append(robotNames, getRobotStatus(statusList, robot.RobotID))
	}
	maxRobot := "无"
	if res.MaxRobotCount > 0 {
		maxRobot = strconv.Itoa(res.MaxRobotCount)
	}
	return areaId, maxRobot, robotNames
}

func getRobotStatus(statusList []dto.RobotStatus, robotId string) string {
	for _, status := range statusList {
		if status.RobotId == robotId {
			return status.RobotName
		}
	}
	log.Errorf("调度资源未知机器人:%s", robotId)
	return "未知机器人"
}

func getPositionName(positions []dao.RobotPosition, positionId string) string {
	for _, position := range positions {
		if position.Id == positionId {
			return position.FullName
		}
	}
	log.Errorf("调度资源未知位置点:%s", positionId)
	return "未知位置点"
}

func sendMsg(msg *baseRedis.Message) {
	switch msg.Channel {
	case constant.WebsocketQueues[0]:
		var data dto.RobotStatus
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		sendAllRobotStatus(data)
	case constant.WebsocketQueues[1]:
		var data model.ElasticRobotStatus
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		sendRobotStatus(data)
	case constant.WebsocketQueues[2]:
		var data []bizMq.RobotJobStatus
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		sendRobotJob(data)
	case constant.WebsocketQueues[3]:
		var data model.ElasticRobotPushMessage
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		sendRobotPush(false, data)
	case constant.WebsocketQueues[4]:
		var data dto.RobotStatus
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		sendPosition(data)
	case constant.WebsocketQueues[5]:
		var data DispatchDTO
		_ = json.Unmarshal([]byte(msg.Payload), &data)
		sendDispatchResource(data.OfficeId, data.NotifyInfo)
	}
}
