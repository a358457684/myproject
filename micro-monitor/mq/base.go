package mq

import (
	"micro-common1/biz/enum"
)

const (
	// 发布调度系统释放资源
	exchangeDispatchRelease   = "exchange_dispatch_release_resource"
	routingKeyDispatchRelease = "routing_key_dispatch_release_resource"

	// 订阅机器人任务变化的队列
	monitorJobChangeQueue = "monitor_job_change_queue"
)

type DispatchVo struct {
	Path enum.DispatchReleaseEnum `json:"path"`
	Data interface{}              `json:"data"`
}
