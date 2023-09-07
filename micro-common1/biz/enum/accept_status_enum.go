package enum

type AcceptStatusEnum int8

const (
	ASUndefind AcceptStatusEnum = iota
	ASWait
	ASCompleted
	ASTimeout
)

func (as AcceptStatusEnum) String() string {
	switch as {
	case ASWait:
		return "待接收"
	case ASCompleted:
		return "已接收"
	case ASTimeout:
		return "超时未取"
	}
	return "未知"
}
