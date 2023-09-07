package enum

// 任务状态
type JobStatusEnum int8

const (
	JsQueue                   JobStatusEnum = iota + 1 // 任务队列中
	JsCalling                                          // 任务下发中
	JsStated                                           // 任务执行中
	JsArrived                                          // 任务到达
	JsCompleted                                        // 任务完成
	JsCancel                                           // 取消任务
	JsLowElectric                                      // 电量低终止
	JsNothingDisinfectant                              // 无消毒液终止
	JsNewTask                                          // 新任务终止
	JsSystemSet                                        // 系统设置结束
	JsCancelResourceNotExists                          // 资源不存在导致取消任务
	JsCancelExpiry                                     // 任务过期取消任务
	JsCreateFail                                       // 任务创建失败
	JsAPPError                                         // APP异常
)

func (t JobStatusEnum) IsEnd() bool {
	return t > JsArrived
}

func (t JobStatusEnum) Code() int {
	return int(t)
}

func (t JobStatusEnum) Message() string {
	switch t {
	case JsQueue:
		return "任务队列中"
	case JsCalling:
		return "任务下发中"
	case JsStated:
		return "任务执行中"
	case JsArrived:
		return "任务到达"
	case JsCompleted:
		return "任务完成"
	case JsCancel:
		return "取消任务"
	case JsLowElectric:
		return "电量低终止"
	case JsNothingDisinfectant:
		return "无消毒液终止"
	case JsNewTask:
		return "新任务终止"
	case JsSystemSet:
		return "系统终止"
	case JsCancelResourceNotExists:
		return "资源缺失"
	case JsCancelExpiry:
		return "任务过期"
	case JsCreateFail:
		return "任务创建失败"
	case JsAPPError:
		return "APP异常"
	default:
		return "未知的任务状态"
	}
}

type SimpleJobStatus struct {
	Code        JobStatusEnum `json:"code"`
	Description string        `json:"description"`
}

func GetAllJobStatus() (statusList []SimpleJobStatus) {
	for i := 1; i <= JsAPPError.Code(); i++ {
		jobStatusEnum := JobStatusEnum(i)
		statusList = append(statusList, SimpleJobStatus{jobStatusEnum, jobStatusEnum.Message()})
	}
	return
}
