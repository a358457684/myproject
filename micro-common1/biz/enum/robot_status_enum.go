package enum

import (
	"strings"
	"unsafe"
)

const unknown = "UNKNOWN"

type RobotStatusEnum int8

// 机器人状态
const (
	RsWaitForDelivery       RobotStatusEnum = iota + 1 // 待配送
	RsCharging                                         // 充电中
	RsOpenDoorForDelivery                              // 开门待配送
	RsDelivering                                       // 配送中
	RsArrivedForReceiver                               // 到达待接收
	RsOpenDoorForReceiver                              // 开门待接收
	RsReturnOrigin                                     // 返程中
	RsWaitForLift                                      // 等待电梯
	RsWalkingIntoLift                                  // 进电梯
	RsInLift                                           // 乘电梯
	RsWalkingOutLift                                   // 出电梯
	RsLockForHandle                                    // 操作中
	RsWaiting                                          // 等待中
	RsIdle                                             // 空闲中
	RsPause                                            // 暂停
	RsFinished                                         // 任务完成
	RsStop                                             // 紧急停止
	RsCalling                                          // 呼叫任务中
	RsOutOfLift                                        // 已出电梯
	RsToCharge                                         // 返回充电中
	RsCancel                                           // 取消任务中
	RsMoveLiftWaiting                                  // 移动电梯等待点中
	RsMoveConflictWaiting                              // 移动冲突等待点中
	RsArriveLiftWaiting                                // 已到达电梯等待点
	RsArriveConflictWaiting                            // 已到达冲突等待点
	RsDisinfectTask                                    // 消毒中  执行消毒任务
	RsDisinfectProcess                                 // 前往消毒中 相当于配送中 消毒过程  去消毒点
	RsMoveBuildingWaiting                              // 去跨楼宇停靠点
	RsArriveBuildingWaiting                            // 已到达跨楼宇停靠点
	RsMoveReplenishPoint                               // 去往补货点 补货提醒
	RsReplenish                                        // 补货中
	RsDispatch                                         // 调度中
	RsMoveWardRound                                    // 去查房中
	RsFailed                RobotStatusEnum = 99       // 任务失败
)

// 是否充电
func (t RobotStatusEnum) IsCharge() bool {
	switch t {
	case RsCharging, RsToCharge:
		return true
	default:
		return false
	}
}

// 是否空闲，待配送，等待中 ,任务完成 都属于空闲状态
func (t RobotStatusEnum) IsIdle() bool {
	switch t {
	case RsFinished, RsIdle, RsWaitForDelivery, RsWaiting, RsCancel, RsFailed:
		return true
	default:
		return false
	}
}

// 是否被呼叫
func (t RobotStatusEnum) IsCalling() bool {
	switch t {
	case RsCalling:
		return true
	default:
		return false
	}
}

// 是否返程
func (t RobotStatusEnum) IsBack() bool {
	switch t {
	case RsReturnOrigin:
		return true
	default:
		return false
	}
}

// 是否消毒
func (t RobotStatusEnum) IsDisinfect() bool {
	switch t {
	case RsDisinfectTask, RsDisinfectProcess:
		return true
	default:
		return false
	}
}

func (t RobotStatusEnum) IsBusy() bool {
	switch t {
	case RsFinished, RsIdle, RsWaitForDelivery, RsCharging:
		return false
	default:
		return true
	}
}

func (t RobotStatusEnum) IsMoving() bool {
	if t == RsPause || t == RsCharging || t == RsWaiting || t == RsStop || t == RsInLift || t == RsWaitForLift || t == RsWalkingIntoLift || t == RsWalkingOutLift ||
		t == RsLockForHandle {
		return false
	}
	// return t == RsDelivering || t == RsReturnOrigin || t == RsToCharge || t == RsMoveBuildingWaiting
	return true
}

func (t RobotStatusEnum) CanCharge() bool {
	switch t {
	case RsIdle, RsCharging:
		return true
	default:
		return false
	}
}

// 不能发送任务
func (t RobotStatusEnum) CanNotSendTask() bool {
	switch t {
	case RsInLift, RsWalkingIntoLift, RsWaitForLift, RsWalkingOutLift, RsIdle:
		return true
	}
	return false
}

// 是否在电梯附近 等电梯或电梯中的状态
func (t RobotStatusEnum) NearLift() bool {
	switch t {
	case RsInLift, RsWaitForLift:
		return true
	}
	return false
}

// 是否与电梯操作相关 包括 等电梯、进电梯、乘电梯、出电梯、已出电梯
func (t RobotStatusEnum) IsLiftStatus() bool {
	switch t {
	case RsWaitForLift, RsInLift, RsOutOfLift, RsWalkingIntoLift, RsWalkingOutLift:
		return true
	}
	return false
}

func (t RobotStatusEnum) Code() int {
	return int(t)
}
func (t RobotStatusEnum) String() string {
	switch t {
	case RsWaitForDelivery:
		return "WAIT_FOR_DELIVERY"
	case RsCharging:
		return "RECHARGEING"
	case RsOpenDoorForDelivery:
		return "OPEN_DOOR_FOR_DELIVERY"
	case RsDelivering:
		return "DELIVERING"
	case RsArrivedForReceiver:
		return "ARRIVERED_FOR_RECEIVER"
	case RsOpenDoorForReceiver:
		return "OPEN_DOOR_FOR_RECEIVER"
	case RsReturnOrigin:
		return "RETURN_ORIGIN"
	case RsWaitForLift:
		return "WAIT_FOR_LIFT"
	case RsWalkingIntoLift:
		return "WALKING_INTO_LIFT"
	case RsInLift:
		return "IN_LIFT"
	case RsWalkingOutLift:
		return "WALKING_OUT_LIFT"
	case RsLockForHandle:
		return "LOCK_FOR_HANDLE"
	case RsWaiting:
		return "WAITING"
	case RsIdle:
		return "IDLE"
	case RsPause:
		return "PAUSE"
	case RsFinished:
		return "FINISHED"
	case RsStop:
		return "STOP"
	case RsCalling:
		return "CALLING"
	case RsOutOfLift:
		return "OUT_OF_LIFT"
	case RsToCharge:
		return "TO_RECHARGE"
	case RsCancel:
		return "CANCEL"
	case RsMoveLiftWaiting:
		return "MOVE_LIFT_WAITING"
	case RsMoveConflictWaiting:
		return "MOVE_CONFLICT_WAITING"
	case RsArriveLiftWaiting:
		return "ARRIVE_LIFT_WAITING"
	case RsArriveConflictWaiting:
		return "ARRIVE_CONFLICT_WAITING"
	case RsDisinfectTask:
		return "DISINFECT_TASK"
	case RsDisinfectProcess:
		return "DISINFECT_PROCESS"
	case RsMoveBuildingWaiting:
		return "MOVE_BUILDING_WAITING"
	case RsArriveBuildingWaiting:
		return "ARRIVE_BUILDING_WAITING"
	case RsMoveReplenishPoint:
		return "MOVE_REPLENISH_POINT"
	case RsReplenish:
		return "REPLENISH"
	case RsDispatch:
		return "DISPATCH"
	case RsFailed:
		return "FAILED"
	case RsMoveWardRound:
		return "MOVE_WARD_ROUND"
	default:
		return unknown
	}
}

func (t RobotStatusEnum) Description() string {
	switch t {
	case RsWaitForDelivery:
		return "待配送"
	case RsCharging:
		return "充电中"
	case RsOpenDoorForDelivery:
		return "开门待配送"
	case RsDelivering:
		return "配送中"
	case RsArrivedForReceiver:
		return "到达待接收"
	case RsOpenDoorForReceiver:
		return "开门待接收"
	case RsReturnOrigin:
		return "返程中"
	case RsWaitForLift:
		return "等待电梯"
	case RsWalkingIntoLift:
		return "进电梯"
	case RsInLift:
		return "乘电梯"
	case RsWalkingOutLift:
		return "出电梯"
	case RsLockForHandle:
		return "操作中"
	case RsWaiting:
		return "等待中"
	case RsIdle:
		return "空闲中"
	case RsPause:
		return "暂停"
	case RsFinished:
		return "任务完成"
	case RsStop:
		return "紧急停止"
	case RsCalling:
		return "呼叫任务中"
	case RsOutOfLift:
		return "已出电梯"
	case RsToCharge:
		return "返回充电中"
	case RsCancel:
		return "取消任务中"
	case RsMoveLiftWaiting:
		return "移动电梯等待点中"
	case RsMoveConflictWaiting:
		return "移动冲突等待点中"
	case RsArriveLiftWaiting:
		return "已到达电梯等待点"
	case RsArriveConflictWaiting:
		return "已到达冲突等待点"
	case RsDisinfectTask:
		return "消毒中" // 执行消毒任务
	case RsDisinfectProcess:
		return "前往消毒中" // 相当于配送中 消毒过程 去消毒点
	case RsMoveBuildingWaiting:
		return "去跨楼宇停靠点"
	case RsArriveBuildingWaiting:
		return "已到达跨楼宇停靠点"
	case RsMoveReplenishPoint:
		return "去往补货点" // 补货提醒
	case RsReplenish:
		return "补货中"
	case RsDispatch:
		return "调度中"
	case RsMoveWardRound:
		return "去查房中"
	case RsFailed:
		return "任务失败"
	default:
		return "未知机器人状态"
	}
}

func RobotStatusEnumByName(statusName string) RobotStatusEnum {
	switch statusName {
	case "ARRIVERED_FOR_RECEIVER":
		return RsArrivedForReceiver
	case "IDLE":
		return RsIdle
	case "DELIVERING":
		return RsDelivering
	case "RECHARGEING":
		return RsCharging
	case "OPEN_DOOR_FOR_DELIVERY":
		return RsOpenDoorForDelivery
	case "OPEN_DOOR_FOR_RECEIVER":
		return RsOpenDoorForReceiver
	case "WAIT_FOR_DELIVERY":
		return RsWaitForDelivery
	case "RETURN_ORIGIN":
		return RsReturnOrigin
	case "WAIT_FOR_LIFT":
		return RsWaitForLift
	case "WALKING_INTO_LIFT":
		return RsWalkingIntoLift
	case "IN_LIFT":
		return RsInLift
	case "OUT_OF_LIFT":
		return RsOutOfLift
	case "WALKING_OUT_LIFT":
		return RsWalkingOutLift
	case "LOCK_FOR_HANDLE":
		return RsLockForHandle
	case "STOP":
		return RsStop
	case "WAITING":
		return RsWaiting
	case "PAUSE":
		return RsPause
	case "FINISHED":
		return RsFinished
	case "CALLING":
		return RsCalling
	case "FAILED":
		return RsFailed
	case "CANCEL":
		return RsCancel
	case "MOVE_LIFT_WAITING":
		return RsMoveLiftWaiting
	case "MOVE_CONFLICT_WAITING":
		return RsMoveConflictWaiting
	case "ARRIVE_LIFT_WAITING":
		return RsArriveLiftWaiting
	case "ARRIVE_CONFLICT_WAITING":
		return RsArriveConflictWaiting
	case "TO_RECHARGE":
		return RsToCharge
	case "REPLENISH":
		return RsReplenish
	case "DISINFECT_TASK":
		return RsDisinfectTask
	case "DISINFECT_PROCESS":
		return RsDisinfectProcess
	case "MOVE_BUILDING_WAITING":
		return RsMoveBuildingWaiting
	case "ARRIVE_BUILDING_WAITING":
		return RsArriveBuildingWaiting
	case "DISPATCH":
		return RsDispatch
	case "MOVE_WARD_ROUND":
		return RsMoveWardRound
	default:
		return 0
	}
}

// 机器人的检查状态集合
type CheckRobotStatusCollection uint64

func (collection CheckRobotStatusCollection) IsChecked(status RobotStatusEnum) bool {
	if status == RsFailed {
		status = 32
	}
	if status < 64 {
		var bf [8]byte
		*(*uint64)(unsafe.Pointer(&bf[0])) = uint64(collection)
		btIndex := status / 8
		offset := status % 8
		value := byte(1 << uint(offset))
		return bf[btIndex]&value == value
	}
	return false
}

// 选中
func (collection *CheckRobotStatusCollection) Checked(status RobotStatusEnum) {
	if status == RsFailed {
		status = 32
	}
	if status < 64 {
		var bf [8]byte
		*(*uint64)(unsafe.Pointer(&bf[0])) = uint64(*collection)
		btIndex := status / 8
		offset := status % 8
		value := byte(1 << uint(offset))
		bf[btIndex] = bf[btIndex] | value
		*collection = *(*CheckRobotStatusCollection)(unsafe.Pointer(&bf[0]))
	}
}

// 非选中
func (collection *CheckRobotStatusCollection) UnChecked(status RobotStatusEnum) {
	if status == RsFailed {
		status = 32
	}
	if status < 64 {
		var bf [8]byte
		*(*uint64)(unsafe.Pointer(&bf[0])) = uint64(*collection)
		btIndex := status / 8
		offset := status % 8
		value := byte(1 << uint(offset))
		value = ^value
		bf[btIndex] = bf[btIndex] & value
		*collection = *(*CheckRobotStatusCollection)(unsafe.Pointer(&bf[0]))
	}
}

// 返回状态字符串
func (collection CheckRobotStatusCollection) String() string {
	resultArrs := make([]string, 0, 32)
	for i := RsWaitForDelivery; i <= RsReplenish; i++ {
		if collection.IsChecked(i) {
			resultArrs = append(resultArrs, i.String())
		}
	}
	if collection.IsChecked(RsFailed) {
		resultArrs = append(resultArrs, RsFailed.String())
	}
	return strings.Join(resultArrs, ",")
}

// 返回状态集合
func (collection CheckRobotStatusCollection) Statuss() []RobotStatusEnum {
	resultArrs := make([]RobotStatusEnum, 0, 32)
	for i := RsWaitForDelivery; i <= RsReplenish; i++ {
		if collection.IsChecked(i) {
			resultArrs = append(resultArrs, i)
		}
	}
	if collection.IsChecked(RsFailed) {
		resultArrs = append(resultArrs, RsFailed)
	}
	return resultArrs
}

// 根据字符串列表重置选择状态
func (collection *CheckRobotStatusCollection) ResetCheckStatus(statusStrings string) {
	status := strings.FieldsFunc(statusStrings, func(r rune) bool {
		return r == ','
	})
	*collection = 0
	for i := 0; i < len(status); i++ {
		RStatus := RobotStatusEnumByName(status[i])
		if RStatus != 0 {
			collection.Checked(RStatus)
		}
	}
}

type SimpleRobotStatus struct {
	Code        RobotStatusEnum `json:"code"`
	Description string          `json:"description"`
}

func GetAllRobotStatus() (statusList []SimpleRobotStatus) {
	for i := 1; i <= RsMoveWardRound.Code(); i++ {
		statusEnum := RobotStatusEnum(i)
		statusList = append(statusList, SimpleRobotStatus{statusEnum, statusEnum.Description()})
	}
	statusList = append(statusList, SimpleRobotStatus{RsFailed, RsFailed.Description()})
	return
}
