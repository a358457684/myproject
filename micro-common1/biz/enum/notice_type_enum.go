package enum

type NoticeTypeEnum int8

const (
	NtAlert NoticeTypeEnum = iota + 1
	NtArrival
	NtAccept
)

func (tp NoticeTypeEnum) Code() int {
	return int(tp)
}

func (tp NoticeTypeEnum) String() string {
	switch tp {
	case NtAlert:
		return "预警通知"
	case NtArrival:
		return "到达通知"
	case NtAccept:
		return "接收通知"
	}
	return "未知通知类型"
}
