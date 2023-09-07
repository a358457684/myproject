package enum

type WarnLevel uint8

const (
	WLUndefind WarnLevel = iota
	WLLow
	WLMedium
	WLHigh
)

func (w WarnLevel) String() string {
	switch w {
	case WLLow:
		return "低"
	case WLMedium:
		return "中"
	case WLHigh:
		return "高"
	default:
		return "未定义"
	}
}
