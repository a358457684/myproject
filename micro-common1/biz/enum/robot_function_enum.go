package enum

//任务结束的状态类型
type RobotFunctionEnum int8

const (
	RfTest      RobotFunctionEnum = iota //测试
	RfCallFirst                          //呼叫优先于返程
	RfDisinfect                          //消毒相关功能
	RfHvc                                //高值耗材相关功能
)

func (r RobotFunctionEnum) Code() int {
	return int(r)
}

func (r RobotFunctionEnum) String() string {
	switch r {
	case RfTest:
		return "测试"
	case RfCallFirst:
		return "呼叫优先于返程"
	case RfDisinfect:
		return "消毒相关功能"
	case RfHvc:
		return "高值耗材相关功能"
	default:
		return ""
	}
}
