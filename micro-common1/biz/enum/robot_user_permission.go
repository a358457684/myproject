package enum

type RobotUserPermissionType string

const (
	RupCallRobotGroupPermission = "device:robot:callJobGroup"
	// 任务查询(收发任务-->只查询自己发送或者接收的任务)权限标识
	RupQueryJobListSendAndReceiveTaskPermission = "device:robotJob:queueList:sendAndReceiveTask"
)

const (
	RuaOffice  RobotUserPermissionType = "10" //机构下机器人
	RuaDEPT    RobotUserPermissionType = "20" //部门下机器人
	RuaAppoint RobotUserPermissionType = "30" //指定机器人
)

func (t RobotUserPermissionType) Code() string {
	return string(t)
}

func (t RobotUserPermissionType) String() string {
	switch t {
	case RuaOffice:
		return "全部机器人"
	case RuaDEPT:
		return "部门机器人"
	case RuaAppoint:
		return "指定机器人"
	default:
		return ""
	}
}
