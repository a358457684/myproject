package enum

type PathTaskTypeEnum int16

//机器人的路径任务类型,目前机器人内部的任务类型，使用的是这个
const (
	PttMove                         PathTaskTypeEnum = iota //移动任务
	PttMoveToElevatorEnterPoint                             //移动任务，移动到电梯口
	PttMoveToElevatorInsidePoint                            //移动任务，移动到电梯内
	PttCallElevator                                         //电梯任务，呼叫电梯
	PttPressFloor                                           //电梯任务，选择楼层
	PttLeaveElevatorExitPoint                               //移动任务，出电梯任务移动到电梯口
	PttMoveToElevatorDest                                   //移动任务，移动到目的地，未使用
	PttWaitReceive                                          //等待任务，到达目的地后，进行等待
	PttSwitchFloorMap                                       //到达电梯内的切换地图任务
	PttMoveToSwitchBuild                                    //移动到楼宇切换点
	PttSwitchBuildMap                                       //到达楼宇切换点，切换楼宇地图任务
	PttAfterCallElevatorWaitComing                          //呼叫电梯后，等待电梯到达
	PttAfterPressElevatorWaitComing                         //电梯按楼层后，等待电梯到达
	PttDisinfect                                            //消毒集合里，每个点位的任务类型
	PttBegin                        PathTaskTypeEnum = 100  //任务开始
	PttEnd                          PathTaskTypeEnum = 101  //任务结束
	PttEnterElevatorFailRollback    PathTaskTypeEnum = 1000 //移动任务，进电梯失败回退点
	PttExitElevatorFailRollback     PathTaskTypeEnum = 1001 //移动任务，出电梯失败回退点
)

func (t PathTaskTypeEnum) String() string {

	switch t {
	case PttMove:
		return "移动任务"
	case PttMoveToElevatorEnterPoint:
		return "移动任务，移动到电梯口"
	case PttMoveToElevatorInsidePoint:
		return "移动任务，移动到电梯内"
	case PttCallElevator:
		return "电梯任务，呼叫电梯"
	case PttPressFloor:
		return "电梯任务，选择楼层"
	case PttLeaveElevatorExitPoint:
		return "移动任务，出电梯任务移动到电梯口"
	case PttMoveToElevatorDest:
		return "移动任务，移动到目的地，未使用"
	case PttWaitReceive:
		return "等待任务，到达目的地后，进行等待"
	case PttSwitchFloorMap:
		return "到达电梯内的切换地图任务"
	case PttMoveToSwitchBuild:
		return "移动到楼宇切换点"
	case PttSwitchBuildMap:
		return "到达楼宇切换点，切换楼宇地图任务"
	case PttAfterCallElevatorWaitComing:
		return "呼叫电梯后，等待电梯到达"
	case PttAfterPressElevatorWaitComing:
		return "电梯按楼层后，等待电梯到达"
	case PttDisinfect:
		return "消毒集合里，每个点位的任务类型"
	case PttBegin:
		return "任务开始"
	case PttEnd:
		return "任务结束"
	case PttEnterElevatorFailRollback:
		return "移动任务，进电梯失败回退点"
	case PttExitElevatorFailRollback:
		return "移动任务，出电梯失败回退点"
	default:
		return "未定义"
	}
}
