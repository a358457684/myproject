package main

import (
	"epshealth-airobot-monitor/api"
	"epshealth-airobot-monitor/cron"
	monitorMqtt "epshealth-airobot-monitor/monitor_mqtt"
	monitorWebsocket "epshealth-airobot-monitor/monitor_websocket"
	"epshealth-airobot-monitor/mq"
	"micro-common1/biz/manager"
	"micro-common1/orm"
)

// @title EPSHealth-AIRobot-Monitor API
// @version 1.0
// @description 物流机器人监控系统API.
// @termsOfService http://epshealth.com/

// @contact.name 联系我们
// @contact.url http://epshealth.com/Contact/contact.html
// @contact.email notice@epsit.cn

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @BasePath /api
func main() {

	// 初始化机器人类型
	_ = manager.InitRobotTypeFromOrm(orm.DB)

	// 订阅RabbitMQ消息
	mq.InitSubscribe()

	// 订阅MQTT消息
	monitorMqtt.InitSubscribe()

	// 初始化定时器
	cron.Init()

	go monitorWebsocket.SubscribeWebSocketQueue()

	api.Init()
}
