package monitor_mqtt

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/service"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"micro-common1/biz/cache"
	"micro-common1/biz/enum"
	bizMq "micro-common1/biz/mq"
	"micro-common1/log"
	"micro-common1/redis"
	"strconv"
	"time"
)

// 离线检测 constant.RobotStatusIntervalTime
func RobotOffLineHandler() {
	if ok := utils.GetLock(constant.MonitorOnlineLock, time.Second*5); !ok {
		return
	}
	robots, err := cache.FindRobotStatusAll()
	if err != nil {
		log.WithError(err).Errorf("没有获取到机器人状态信息")
		return
	}
	for _, robot := range robots {
		if robot.NetStatus == enum.NsOffline {
			continue
		}
		// 心跳包时间间隔大于设定值，设置为离线
		intervalTime := time.Now().Unix() - robot.Time.Unix()
		if intervalTime > constant.RobotStatusIntervalTime {
			oldStatus := robot
			id := robot.RobotId
			log.Infof("机器人<%s,%s>超过%d秒没有上传状态", id, robot.RobotName, constant.RobotStatusIntervalTime)
			robot.NetStatus = enum.NsOffline
			robot.Time = time.Now()
			// 离线后处理在线时间
			onLineTimeHandler(id)
			err = cache.SaveRobotStatus(robot.OfficeId, id, robot)
			if err != nil {
				log.WithError(err).Errorf("设置机器人:%s，离线失败", id)
				continue
			}
			log.Infof("====== robotId<%s,%s>; disconnected ======", id, robot.RobotName)
			// 离线预警
			setMonitorNetConnect(robot.OfficeId, id, enum.NsOffline)
			socketData, _ := json.Marshal(robot)
			redis.Publish(context.Background(), constant.WebsocketQueues[0], socketData)
			service.AddRobotStatusDocument(robot, fmt.Sprintf("系统检测设置为离线，最后上传状态时间：%s",
				oldStatus.Time.Format(constant.DateTimeFormat)))
			// 给Y2P发送离线状态
			for _, model := range constant.RobotPads {
				if string(robot.RobotModel) == model {
					publishRobotToPad(robot)
					break
				}
			}
			// 给Y2R、E2R发送离线状态
			sendRobotStatusToMobileTerminal(robot)
			_ = bizMq.RobotStatusUpdate(bizMq.RobotStatusChangeMqVo{
				OldStatus: oldStatus,
				NewStatus: robot,
				SentTime:  time.Now().UnixNano() / 1e6,
			})
		}
	}
}

// 离线后处理在线时间
func onLineTimeHandler(id string) {
	ctx := context.Background()
	onLineKey := constant.LogisticsRobotOnline + id
	onLineStr, err := redis.Get(ctx, onLineKey).Result()
	if err != nil {
		log.WithError(err).Errorf("处理在线时间失败")
		return
	}
	onLineNano, convertErr := strconv.ParseInt(onLineStr, 10, 64)
	if convertErr != nil {
		log.WithError(err).Errorf("转换在线时间失败")
	}
	timeInterval := time.Now().UnixNano() - onLineNano
	onLineTimeKey := constant.LogisticsRobotOnlineTime + id
	onLineTimeStr, err := redis.Get(ctx, onLineTimeKey).Result()
	onLineTimeNano, convertErr := strconv.ParseInt(onLineTimeStr, 10, 64)
	if convertErr != nil {
		log.WithError(err).Errorf("转换在线时间失败")
	}
	if err == nil {
		timeInterval = onLineTimeNano + timeInterval
	}
	redis.Set(ctx, onLineTimeKey, timeInterval, 0)
	redis.Del(ctx, onLineKey)
}
