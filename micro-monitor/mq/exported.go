package mq

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	monitorWebsocket "epshealth-airobot-monitor/monitor_websocket"
	"micro-common1/biz/enum"
	bizMq "micro-common1/biz/mq"
	"micro-common1/log"
	"micro-common1/rabbitmq"
	"micro-common1/redis"
	"time"
)

// 初始化订阅的RabbitMQ
func InitSubscribe() {
	subscribeDispatchResource()
	subscribeRobotJob()
}

// 订阅：调度资源占用情况
func subscribeDispatchResource() {
	_ = bizMq.SubResUseChangeNotify(func(officeId string, notifyInfo ...bizMq.ResourceNotify) {
		log.Infof("调度信息: %s, %+v", officeId, notifyInfo)
		socketData, _ := json.Marshal(monitorWebsocket.DispatchDTO{OfficeId: officeId, NotifyInfo: notifyInfo})
		redis.Publish(context.Background(), constant.WebsocketQueues[5], socketData)
	})
}

// 订阅：任务发生变化
func subscribeRobotJob() {
	_ = bizMq.SubscribeRobotJobStatus(monitorJobChangeQueue, func(time time.Time, data ...bizMq.RobotJobStatus) {
		socketData, _ := json.Marshal(data)
		redis.Publish(context.Background(), constant.WebsocketQueues[2], socketData)
		jobStatusCheck(data)
	})
}

// 发布：消息到调度系统，释放资源
func PublishDispatch(path enum.DispatchReleaseEnum, data interface{}) error {
	log.Infof("=====发送消息到调度系统: 类型：%s, 数据：%+v", path.String(), data)
	err := rabbitmq.Publish(
		rabbitmq.Exchange{
			Name:  exchangeDispatchRelease,
			Model: rabbitmq.ET_Direct,
		},
		routingKeyDispatchRelease,
		rabbitmq.NewMessage(DispatchVo{path, data}),
	)
	if err != nil {
		log.WithError(err).Error("发送RabbitMQ消息失败")
	}
	return err
}
