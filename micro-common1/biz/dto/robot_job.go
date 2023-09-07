package dto

import (
	"common/biz/enum"
	"time"
)

//任务组Vo
type ApplyRobotJob struct {
	Origin   enum.MsgOriginEnum //消息来源
	JobType  enum.JobTypeEnum   //任务类型
	Back     bool               //是否立即返回
	OfficeID string             //机构ID
	RobotID  string             //机器人ID，为空时由调度系统来分配机器人
	GroupID  string             //任务组ID
	UserID   string             //发起任务的用户ID
	Order    int                //任务优先级
	Jobs     []JobItem          //任务集合
	Time     time.Time          //发起时间
}

//任务Vo
type JobItem struct {
	Floor     int    //楼层
	JobID     string `json:"JobID,omitempty"`   //任务ID
	BuildID   string `json:"BuildID,omitempty"` //楼宇ID
	BuildName string `json:"BuildName,omitempty"`
	GUID      string `json:"GUID,omitempty"`    //位置GUID
	PosName   string `json:"PosName,omitempty"` //位置GUID
}

//任务信息
type BaseRobotJob struct {
	Origin   enum.MsgOriginEnum //消息来源
	OfficeID string             //机构ID
	RobotID  string             //机器人ID
	GroupID  string             //任务组ID
	JobID    string             //任务ID,为空时表示取消整个任务组
}

//任务到达
type RobotJobArrived struct {
	BaseRobotJob

	ArrivedTime time.Time
}

//任务完成/取消
type RobotJobCompleted struct {
	BaseRobotJob

	Status          enum.JobStatusEnum    //状态
	CompletedTime   time.Time             //完成时间
	CompletedUserID string                //完成者
	Distance        int                   //行驶里程
	AcceptState     enum.AcceptStatusEnum //物品接收状态
	Remarks         string                //备注，失败/取消原因
}

//消毒任务
type ApplyDisinfectJob struct {
	JobType         enum.JobTypeEnum   //任务类型
	Origin          enum.MsgOriginEnum //消息来源
	OfficeID        string             //机构ID
	RobotID         string             //机器人ID，为空时由调度系统来分配机器人
	GroupID         string             //任务组ID
	JobID           string             //任务ID
	UserID          string             //发起任务的用户ID
	TaskID          string             //定时消毒任务Id
	TaskName        string             //任务名称
	Areas           []DisinfectArea    //消毒区域集合
	EchoTime        string             //重复时间
	PlanConsumeTime string             //预计耗时(分钟)
	Order           int                //任务优先级
	Back            bool               //是否立即返回
	Time            time.Time          //发起时间
}

//消毒区域
type DisinfectArea struct {
	ID              string              //消毒区域ID
	Name            string              //消毒区域名称
	BuildingId      string              //楼宇ID
	Floor           string              //楼层
	PlanConsumeTime string              //预计耗时(分钟)
	SpraySize       int                 //喷雾量大小设置0:小；1:中；2:大
	Positions       []DisinfectPosition //位置点集合
}

//消毒点位
type DisinfectPosition struct {
	GuId          string //guId
	Name          string //位置名称
	DisinfectTime string //消毒时间
}
