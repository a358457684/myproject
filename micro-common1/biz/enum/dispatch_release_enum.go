package enum

// 通知调度系统释放资源
type DispatchReleaseEnum int8

const (
	ReleaseAllResource DispatchReleaseEnum = iota + 1
	CancelAllJob
	ReleaseCacheJob
)

func (tp DispatchReleaseEnum) Code() int {
	return int(tp)
}

func (tp DispatchReleaseEnum) String() string {
	switch tp {
	case ReleaseAllResource:
		return "释放所有资源"
	case CancelAllJob:
		return "取消所有任务"
	case ReleaseCacheJob:
		return "移除缓存任务"
	}
	return "未知操作"
}
