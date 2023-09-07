package enum

//机器人的点位类型
type PositionTypeEnum int16

const (
	PtUndefined          PositionTypeEnum = iota
	PtInit                                //机器初始点
	PtCharge                              //机器充电点
	PtNormal                              //普通任务点
	PtChangeFloor                         //楼层切换点 在电梯内楼层（地图）切换任务
	PtChangeBuilding                      //楼宇切换点 到楼宇切换点任务
	PtWait                                //常规等待点
	PtElevatorWait                        //电梯等待点
	PtEnterElevator                       //进梯任务点，电梯口
	PtEnterElevatorError                  //进梯回退点
	PtOutElevator                         //出梯任务点,去除电梯口
	PtOutElevatorError                    //出梯回退点
	PtDisinfect                           //消毒任务点
)

//是否可以用来在前端展示的点位
func (t PositionTypeEnum) IsCanShow() bool {
	return t == PtNormal || t == PtCharge || t == PtInit
}

//是否初始点
func (t PositionTypeEnum) IsInit() bool {
	return t == PtInit
}

//是否充点电
func (t PositionTypeEnum) IsCharge() bool {
	return t == PtCharge
}

//是否普通点
func (t PositionTypeEnum) IsNormal() bool {
	return t == PtNormal
}

func (t PositionTypeEnum) Code() int16 {
	return int16(t)
}

func (t PositionTypeEnum) String() string {
	switch t {
	case PtInit:
		return "机器初始点"
	case PtCharge:
		return "机器充电点"
	case PtNormal:
		return "常规任务点"
	case PtChangeFloor:
		return "楼层切换点"
	case PtChangeBuilding:
		return "楼宇切换点"
	case PtWait:
		return "常规等待点"
	case PtElevatorWait:
		return "电梯等待点"
	case PtEnterElevator:
		return "进梯任务点"
	case PtEnterElevatorError:
		return "进梯回退点"
	case PtOutElevator:
		return "出梯任务点"
	case PtOutElevatorError:
		return "出梯回退点"
	case PtDisinfect:
		return "消毒任务点"
	default:
		return "未知类型"
	}
}
