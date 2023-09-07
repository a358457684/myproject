package enum

// 任务执行状态
type ExecStateEnum int8

const (
	EsNormal ExecStateEnum = iota + 1
	EsCancel
	EsFailed ExecStateEnum = 99
)

func (tp ExecStateEnum) String() string {
	switch tp {
	case EsNormal:
		return "正常"
	case EsCancel:
		return "取消"
	case EsFailed:
		return "异常"
	}
	return "未知状态"
}
