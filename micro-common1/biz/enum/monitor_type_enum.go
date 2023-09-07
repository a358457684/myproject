package enum

type MonitorTypeEnum string

const (
	MtJobMonitorType   MonitorTypeEnum = "1"
	MtScopeMonitorType MonitorTypeEnum = "2"
)

func (as MonitorTypeEnum) Description() string {
	switch as {
	case MtJobMonitorType:
		return "任务监控"
	case MtScopeMonitorType:
		return "范围监控"
	}
	return "未知"
}
