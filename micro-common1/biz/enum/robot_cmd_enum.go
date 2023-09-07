package enum

type RobotCmdEnum string

// 机器人状态
const (
	RCNone         RobotCmdEnum = "none"
	RCFeedBack     RobotCmdEnum = "feedBack"     // 消息确认
	RCRobotStatus  RobotCmdEnum = "robotStatus"  // 机器人上传状态
	RCRestartRobot RobotCmdEnum = "restartRobot" // 重启底盘
	RCCloseAPP     RobotCmdEnum = "closeAPP"     // 关闭APP
	RCRestartAPP   RobotCmdEnum = "restartAPP"   // 重启APP
	RCCancelEStop  RobotCmdEnum = "cancelEStop"  // 取消急停
	RCUploadLog    RobotCmdEnum = "uploadLog"    // 上传日志
	RCPause        RobotCmdEnum = "pause"        // 调度暂停机器人
	RCCancelPause  RobotCmdEnum = "cancelPause"  // 调度唤醒机器人
	RCCheckSelf    RobotCmdEnum = "checkSelf"    // 当在移动中的机器人，连续多次传递上来的状态和坐标位置没怎么变动的时候下发，机器人需要自己去检查判定
	RCToLift       RobotCmdEnum = "toLift"       // 机器人在去电梯口的过程中，根据电梯资源情况，调度机器人去本楼可以使用的其他电梯资源位置
	RCJob          RobotCmdEnum = "job"          // 调度发送任务,里面包含有呼叫任务
	// RCBack          RobotCmdEnum = "back"          //调度机器人去返程，和dispCmdPause的参数一致,如果指定了返程点，就去指定点，如果没指定就去默认点，返程通过Job任务来，取消这个
	RCCancelJob     RobotCmdEnum = "cancelJob"     // 取消机器人任务，一般是取消一组任务,
	RCOpenUVLamp    RobotCmdEnum = "openUVLamp"    // 打开紫外灯
	RCCloseUVLamp   RobotCmdEnum = "closeUVLamp"   // 关闭紫外灯
	RCOpenAtomizer  RobotCmdEnum = "openAtomizer"  // 打开雾化器
	RCCloseAtomizer RobotCmdEnum = "closeAtomizer" // 关闭雾化器
	RCProxyStatus   RobotCmdEnum = "proxyStatus"   // 代理服务状态
)

func (r RobotCmdEnum) Code() string {
	return string(r)
}

func (r RobotCmdEnum) String() string {
	switch r {
	case RCNone:
		return "空的命令"
	case RCFeedBack:
		return "消息确认"
	case RCRobotStatus:
		return "机器人状态"
	case RCRestartRobot:
		return "重启底盘"
	case RCCloseAPP:
		return "关闭APP"
	case RCRestartAPP:
		return "重启APP"
	case RCCancelEStop:
		return "取消急停"
	case RCUploadLog:
		return "上传日志"
	case RCPause:
		return "机器暂停"
	case RCCancelPause:
		return "取消暂停"
	case RCCheckSelf:
		return "机器自检"
	case RCToLift:
		return "电梯调度"
	case RCJob:
		return "任务下发"
	/*case RCBack:
	return "返程任务"*/
	case RCCancelJob:
		return "取消任务"
	case RCOpenUVLamp:
		return "打开紫外灯"
	case RCCloseUVLamp:
		return "关闭紫外灯"
	case RCOpenAtomizer:
		return "打开雾化器"
	case RCCloseAtomizer:
		return "关闭雾化器"
	case RCProxyStatus:
		return "代理服务状态"
	default:
		return "unknown"
	}
}
