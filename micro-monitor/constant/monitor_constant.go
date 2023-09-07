package constant

import (
	"strings"
)

const (
	// 时间格式
	DateTimeFormat = "2006-01-02 15:04:05"

	// 监控系统权限前缀
	MonitorPermission = "monitor:"

	// 操作权限
	MonitorOperatePermission = "monitor:operate"

	// 调度权限
	MonitorDispatchPermission = "monitor:dispatch"

	// 默认的楼宇
	DefaultBuildingHalt = "A"

	// 机器人状态心跳包超过15秒没上传判断为离线
	RobotStatusIntervalTime = 15

	// 机构配置是否调度记录时间间隔（秒）
	OfficeConfigLogTimespan = 60 * 30

	// 管理员
	AdminName = "管理员"

	// 医护端的管理员用户id
	UserAdminId = "123456789012345678901234567890ab"

	// 医护端系统别名
	UserSystemName = "user"
)

var WebsocketQueues = []string{
	"monitor_websocket_robot_queue",
	"monitor_websocket_status_queue",
	"monitor_websocket_job_queue",
	"monitor_websocket_push_queue",
	"monitor_websocket_monitor_queue",
	"monitor_websocket_dispatch_queue",
}

// 消毒机器人类型
var DisinfectRobots = [2]string{"A1", "A2"}

// pad的机器人类型
var PadRobots = [2]string{"Y2R", "E2R"}

// 目前只有Y2P开放式箱体用到这种
var RobotPads = [1]string{"Y2P"}

type PermissionType string

const (
	/**
	 * 监控系统操作按钮
	 */
	PtCancelRobotJob  PermissionType = "/api/robotJobQueue/cancelRobotJob"
	PtGetRobotStatus  PermissionType = "/api/device/getRobotStatus"
	PtOperateRobot    PermissionType = "/api/robotUser/operateRobot"
	PtRemoveRobot     PermissionType = "/api/robotUser/remove"
	PtDispatchOperate PermissionType = "/api/robotJob/dispatchOperate"
)

func (tp PermissionType) Code() string {
	return string(tp)
}

func (tp PermissionType) String() string {
	switch tp {
	case PtCancelRobotJob:
		return "监控系统-取消任务"
	case PtGetRobotStatus:
		return "监控系统-查看详情"
	case PtOperateRobot:
		return "监控系统-机器操作"
	case PtRemoveRobot:
		return "监控系统-移除机器人"
	case PtDispatchOperate:
		return "监控系统-调度操作"
	default:
		return "监控系统-未知请求"
	}
}

func GetPermission(uri string) string {
	switch PermissionType(uri) {
	case PtCancelRobotJob, PtGetRobotStatus, PtOperateRobot:
		return MonitorOperatePermission
	case PtDispatchOperate:
		return MonitorDispatchPermission
	default:
		if strings.HasPrefix(uri, string(PtRemoveRobot)) {
			return MonitorOperatePermission
		}
		return "未知请求"
	}
}

// 推送状态
type PushStatus int8

const (
	PushSucceed PushStatus = iota + 1
	ExecuteSucceed
	ExecuteFail
)

func (tp PushStatus) String() string {
	switch tp {
	case PushSucceed:
		return "推送成功"
	case ExecuteSucceed:
		return "执行成功"
	case ExecuteFail:
		return "执行失败"
	}
	return "未知状态"
}

type RobotOperationEnum int8

// 操作机器人
const (
	RORestartRobot RobotOperationEnum = iota + 1 // 重启底盘
	ROCloseAPP                                   // 关闭APP
	RORestartAPP                                 // 重启APP
	RCCancelEStop                                // 取消急停
	ROUploadLog                                  // 上传日志
)

func (t RobotOperationEnum) Code() string {
	switch t {
	case RORestartRobot:
		return "restartRobot"
	case ROCloseAPP:
		return "closeAPP"
	case RORestartAPP:
		return "restartAPP"
	case RCCancelEStop:
		return "cancelEStop"
	case ROUploadLog:
		return "uploadLog"
	default:
		return "unknown"
	}
}

func (t RobotOperationEnum) String() string {
	switch t {
	case RORestartRobot:
		return "重启机器人"
	case ROCloseAPP:
		return "退出APP"
	case RORestartAPP:
		return "重启APP"
	case RCCancelEStop:
		return "取消急停"
	case ROUploadLog:
		return "日志上传"
	default:
		return "unknown"
	}
}
