package enum

import (
	"fmt"
	"time"
)

// 自定义任务格式
type CustomJobInfo struct {
	Origin      MsgOriginEnum // 消息来源
	WrapJobType JobTypeEnum   // 包装的标准任务类型，除了查房任务之外的其他任务
	Back        bool          // 是否立即返回
	OfficeID    string        // 机构ID
	RobotID     string        // 机器人ID，为空时由调度系统来分配机器人
	UserID      string        // 发起任务的用户ID
	Order       int           // 任务优先级
	Remark      string        // 备注描述信息
	JobID       string        // 任务ID
	GUID        string        // 任务要到达的位置
	Time        time.Time     // 发起时间
}

// 任务类型
type JobTypeEnum int

const (
	JtNoJob              JobTypeEnum = iota * 10 // 无任务
	JtDistribute                                 // 配送任务
	JtCall                                       // 呼叫任务
	JtBack                                       // 返程任务
	JtWait                                       // 等待任务
	JtCharge                                     // 充电任务
	JtLiftWait                                   // 电梯等待区域任务
	JtControlWait                                // 管制等待区域任务
	JtDisinfect                                  // 定时消毒任务
	JtBuildingWait                               // 到跨楼宇等待位置任务
	JtDisinfectImmediate                         // 即时消毒任务
	JtRounds                                     // 查房任务
)

const (
	// 所有的自定义任务的最高位为1，比如，自定义一个专门的移动到某地的任务，任务ID为
	// 0x8000 | JtCall     //包装了一个呼叫任务到自定义任务，这种任务的主要目的是为了使用机器人的行走,同时记录任务日志到数据库
	JtCustomJobMask = 0x8000 // 自定义任务标记的掩码，这种自定义任务，一次只能到一个位置坐标点，比如呼叫，最高位为1
	JtStandJobMask  = 0x7fff // 标准任务的掩码
)

func (t JobTypeEnum) IsDisinfectJob() bool {
	return t == JtDisinfect || t == JtDisinfectImmediate
}

func (t JobTypeEnum) Code() int {
	return int(t)
}

func (t JobTypeEnum) String() string {
	switch t {
	case JtNoJob:
		return "无任务"
	case JtDistribute:
		return "配送任务"
	case JtCall:
		return "呼叫任务"
	case JtBack:
		return "返程任务"
	case JtWait:
		return "等待任务"
	case JtCharge:
		return "充电任务"
	case JtLiftWait:
		return "电梯等待区域任务"
	case JtControlWait:
		return "交通管制等待区域任务"
	case JtDisinfect:
		return "定时消毒任务"
	case JtBuildingWait:
		return "到跨楼宇等待位置任务"
	case JtDisinfectImmediate:
		return "即时消毒任务"
	case JtRounds:
		return "查房任务"
	default:
		if int(t)&JtCustomJobMask == JtCustomJobMask {
			return "自定义包装" + (t & JtStandJobMask).String()
		}
		return fmt.Sprintf("未知任务类型(%d)", t.Code())
	}
}

type SimpleJobType struct {
	Code        JobTypeEnum `json:"code"`
	Description string      `json:"description"`
}

func GetAllJobType() (statusList []SimpleJobType) {
	for i := 10; i <= JtRounds.Code(); i += 10 {
		jobTypeEnum := JobTypeEnum(i)
		statusList = append(statusList, SimpleJobType{jobTypeEnum, jobTypeEnum.String()})
	}
	return
}
