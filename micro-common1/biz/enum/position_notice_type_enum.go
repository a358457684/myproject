package enum

type PositionNoticeTypeEnum uint8

const (
	PNTInvalid PositionNoticeTypeEnum = iota
	PNTInsideTel
	PNTSoundLight
	PNTTel
	PNTVoice
	PNTWeChat
)

func (as PositionNoticeTypeEnum) String() string {
	switch as {
	case PNTInvalid:
		return "不通知"
	case PNTInsideTel:
		return "内线电话"
	case PNTSoundLight:
		return "声光提醒"
	case PNTTel:
		return "电话通知"
	case PNTVoice:
		return "语音提醒"
	case PNTWeChat:
		return "维信通知"
	}
	return "未知"
}
