package monitor_mqtt

import (
	"epshealth-airobot-monitor/result"
	"fmt"
	mqttBase "github.com/eclipse/paho.mqtt.golang"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	"micro-common1/log"
	"micro-common1/mqtt"
	"micro-common1/util"
	"time"
)

// 初始化订阅的MQTT
func InitSubscribe() {
	subscribeToServer()
	subscribeToClient()
	subscribeToProxy()
}

// 订阅：机器人发送给服务端
func subscribeToServer() {
	err := mqtt.SubscribeCustom(
		func(client *mqtt.Client, data interface{}, message mqttBase.Message) {
			go robotStatusMessageHandler(message.Topic(), data)
		},
		MqttMsgVo{},
		2,
		toServer,
	)
	if err != nil {
		panic(err)
	}
}

// 订阅：服务端发送给机器人的数据，存储推送信息
func subscribeToClient() {
	err := mqtt.SubscribeCustom(
		func(client *mqtt.Client, data interface{}, message mqttBase.Message) {
			go robotPushMessageHandler(message.Topic(), data)
		},
		MqttMsgVo{},
		2,
		fmt.Sprintf(toClient, "+", "+", "+"),
	)
	if err != nil {
		panic(err)
	}
}

// 订阅：代理（工控机、代理服务、电梯等）的状态
func subscribeToProxy() {
	err := mqtt.SubscribeCustom(
		func(client *mqtt.Client, data interface{}, message mqttBase.Message) {
			go proxyMonitorHandler(message.Topic(), data.(*MqttMsgVo))
		},
		MqttMsgVo{},
		2,
		fmt.Sprintf(toProxy, "+", "+", "proxyStatus"),
	)
	if err != nil {
		panic(err)
	}
}

// 发布：token验证失败的返回
func publishInvalidToken(topics []string, msgId int64) {
	data := feedback{
		Cmd:    topics[5],
		MsgID:  msgId,
		Status: result.InvalidToken.Code(),
	}
	err := mqtt.PublishCustom(
		fmt.Sprintf(toClient, topics[3], topics[4], enum.RCFeedBack.Code()),
		2,
		false,
		data)
	if err != nil {
		log.WithError(err).Error("发送MQTT消息失败")
	}
}

// 发布：给机器人pad端发送消息（Y2R、E2R）
func publishPadToOffice(officeId string, code int, data interface{}) {
	mqttMessageData := MqttMessageData{
		Code:     code,
		MqttType: enum.RCRobotStatus.Code(),
		MsgId:    util.CreateUUID(),
		Time:     time.Now().UnixNano() / 1e6,
		Data:     data,
	}
	err := mqtt.PublishCustom(
		padToOffice+officeId,
		2,
		false,
		mqttMessageData)
	if err != nil {
		log.WithError(err).Error("发送MQTT消息失败")
	}
}

// 发布：给机器人pad端发送消息（Y2P）
func publishRobotToPad(robot dto.RobotStatus) {
	path := fmt.Sprintf(toPad, robot.OfficeId, robot.RobotId, util.CreateUUID())
	msg := serverMsgVo{
		Path:  path,
		MsgId: util.CreateUUID(),
		Body:  robot,
	}
	err := mqtt.PublishCustom(
		path,
		2,
		false,
		msg)
	if err != nil {
		log.WithError(err).Error("发送MQTT消息失败")
	}
}
