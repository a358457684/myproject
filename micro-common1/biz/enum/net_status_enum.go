package enum

//网络状态
type NetStatusEnum int

const (
	NsOnline  NetStatusEnum = iota + 1 //在线
	NsOffline                          //离线
)

func (t NetStatusEnum) Code() int {
	return int(t)
}

func (t NetStatusEnum) String() string {
	switch t {
	case NsOnline:
		return "CONNECT"
	case NsOffline:
		return "DISCONNECT"
	default:
		return ""
	}
}

func (t NetStatusEnum) Description() string {
	switch t {
	case NsOnline:
		return "在线"
	case NsOffline:
		return "离线"
	default:
		return "未知网络状态"
	}
}
