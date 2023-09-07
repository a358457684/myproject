package model

import (
	"micro-common1/biz/enum"
	"time"
)

type MonitorBaseInfo struct {
	StartTime time.Time // 开始时间
	OfficeId  string    // 机构ID
	RobotId   string    // 机器人ID
	PushCount int       // 已推次数
}

type JobMonitorVo struct {
	JobId  string
	Status enum.RobotStatusEnum
	MonitorBaseInfo
}

type ByBatteryMonitorVo struct {
	Electric float64
	MonitorBaseInfo
}

type MonitorNetConnectVo struct {
	NetStatus enum.NetStatusEnum // 状态
	MonitorBaseInfo
}

type MonitorStatusVo struct {
	Status enum.RobotStatusEnum // 状态
	MonitorBaseInfo
}

type ScopeMonitorVo struct {
	X      float64
	Y      float64
	Status enum.RobotStatusEnum
	MonitorBaseInfo
}

type MonitorConfig struct {
	Id                string
	OfficeId          string // 机构
	RobotStatus       string // 机器人状态，需要进行定期检测的状态值
	ErrorTimeSpan     int64  // 异常状态时间间隔，即当机器人一直处于robot_status状态时，超过error_time_span时间间隔后，认为是异常状态。
	StatusDescription string //  status_description
	ErrorPushCount    int    // 异常推送次数
	AlertType         string // 预警级别（低：0、中：1、高：2）默认中级
}

type JobScopeMonitorConfig struct {
	Id                string
	OfficeId          string // 机构
	RobotId           string // 机器人ID
	RobotName         string // 机器人名称
	RobotStatus       string // 机器人状态，需要进行定期检测的状态值。
	StatusDescription string // 机器人状态描述名称
	ErrorTimeSpan     int64  // 异常状态时间间隔，即当机器人一直处于robot_status状态时，超过error_time_span时间间隔后，认为是异常状态。
	ErrorPushCount    int    // 异常推送次数
	AlertType         string // 预警级别（低：0、中：1、高：2）默认中级
	MonitorType       string // 监控类型 1.任务检测 2.范围检测
	MonitorScope      string // 监控范围（米）
}
