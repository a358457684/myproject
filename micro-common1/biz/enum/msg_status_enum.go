package enum

type MsgStatusEnum int8

const (
	MsFailed    MsgStatusEnum = iota + 1 //发送失败
	MsSend                               //已发送
	MsConfirmed                          //已确认
)
