package cron

import (
	monitorMqtt "epshealth-airobot-monitor/monitor_mqtt"
	"micro-common1/log"
)

func Init() {
	cronTab := cron.New(cron.WithSeconds())
	checkRobotStatusOnline(cronTab)
	cronTab.Start()
}

// 每30秒检查一次机器人状态，判断是否离线
func checkRobotStatusOnline(cronTab *cron.Cron) {
	spec := "0/30 * * * * ?"
	_, err := cronTab.AddFunc(spec, monitorMqtt.RobotOffLineHandler)
	if err != nil {
		log.WithError(err).Error("Cron mission creation failed!")
	}
}
