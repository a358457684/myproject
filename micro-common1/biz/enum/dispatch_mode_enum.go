package enum

type DispatchModeEnum int8

const (
	DmUndefind DispatchModeEnum = iota
	DmDispatch
	DmStandalone
)

func (tp DispatchModeEnum) String() string {
	switch tp {
	case DmStandalone:
		return "单机模式"
	case DmDispatch:
		return "调度模式"
	}
	return "未知模式"
}

func (tp DispatchModeEnum) IsDispatch() bool {
	return tp == DmDispatch
}
