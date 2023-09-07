package monitor_mqtt

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/model"
	monitorWebsocket "epshealth-airobot-monitor/monitor_websocket"
	"epshealth-airobot-monitor/service"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"micro-common1/biz/enum"
	"micro-common1/log"
	"micro-common1/redis"
	"micro-common1/util"
	"strconv"
	"strings"
	"time"
)

func robotPushMessageHandler(topic string, data interface{}) {
	msg := data.(*MqttMsgVo)
	if ok := utils.GetLock(fmt.Sprintf("%s%d", constant.MonitorPushMessageLock, msg.MsgID), time.Second*3); !ok {
		return
	}

	topics := strings.Split(topic, "/")
	path := topics[5]
	if enum.RobotCmdEnum(path) == enum.RCFeedBack {
		return
	}

	now := time.Now()
	ctx := context.Background()

	// 阻塞, 3秒内收到回传修改状态
	status := constant.PushSucceed
	for i := 0; i < 6; i++ {
		res, err := redis.Get(ctx, fmt.Sprintf("%s%d", constant.PushMessageCallback, msg.MsgID)).Result()
		if err == nil {
			if res == feedbackSucceed {
				status = constant.ExecuteSucceed
			} else {
				status = constant.ExecuteFail
			}
			break
		}
		time.Sleep(time.Millisecond * 500)
	}

	// 重试
	msgRecordKey := fmt.Sprintf("%s%d", constant.PushMessageRecord, msg.MsgID)
	recordData, err := redis.Get(ctx, msgRecordKey).Result()
	if err == nil {
		if dataList := strings.Split(recordData, ","); len(dataList) == 3 {
			sendCount, _ := strconv.Atoi(dataList[2])
			sendCount = sendCount + 1
			if status == constant.PushSucceed {
				dataList[2] = strconv.Itoa(sendCount)
				redis.Set(ctx, msgRecordKey, strings.Join(dataList, ","), time.Minute)
			}
			err = service.UpdatePushMessageSendCount(dataList[0], dataList[1], status, sendCount, now)
			if err == nil {
				monitorWebsocket.SendRobotPush(true, model.ElasticRobotPushMessage{
					OfficeId:   topics[3],
					RobotId:    topics[4],
					DocumentId: dataList[0],
					Status:     status,
					StatusText: status.String(),
					SendCount:  sendCount,
				})
			}
		}
		log.Warnf("消息:%s,重发了:%s", msg.MsgID, recordData)
		return
	}

	// 第一次
	index := fmt.Sprintf("%s-%s", service.RobotPushMessageIndex, now.Format(service.DataFormat))
	message := model.ElasticRobotPushMessage{
		DocumentId:     util.CreateUUID(),
		OfficeId:       topics[3],
		RobotId:        topics[4],
		Path:           path,
		MsgId:          strconv.FormatInt(msg.MsgID, 10),
		Status:         status,
		StatusText:     status.String(),
		FirstTimestamp: now,
		Timestamp:      now,
		SendCount:      1,
	}
	redis.Set(ctx, msgRecordKey, fmt.Sprintf("%s,%s,%d", message.DocumentId, index, 1), time.Minute)
	jsonData, _ := json.Marshal(msg)
	message.Body = string(jsonData)
	service.AddRobotPushMessageDocument(message, index)
	socketData, _ := json.Marshal(message)
	redis.Publish(ctx, constant.WebsocketQueues[3], socketData)
}

func feedbackHandler(feedbackMsg *MqttMsgVo) {
	log.Infof("robot cmd feedback:%+v", feedbackMsg)
	redisKey := fmt.Sprintf("%s%d", constant.PushMessageCallback, feedbackMsg.MsgID)
	redis.Set(context.Background(), redisKey, feedbackMsg.Status, time.Second*10)
}
