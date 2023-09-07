package model

import (
	"gorm.io/gorm"
	"micro-common1/biz/enum"
	"micro-common1/util"
	"time"
)

func (RobotJob) TableName() string {
	return "device_robot_job"
}

func (e *RobotJob) BeforeCreate(_ *gorm.DB) (err error) {
	e.Id = util.CreateUUID()
	e.CreateDate = time.Now()
	e.UpdateDate = time.Now()
	e.DelFlag = DelFlagNormal // 代表正常数据
	return
}

func (e *RobotJob) BeforeUpdate(_ *gorm.DB) (err error) {
	e.UpdateDate = time.Now()
	return
}

type RobotJob struct {
	Id            string
	RobotId       string             // 机器人
	OfficeId      string             // 机构
	JobType       enum.JobTypeEnum   `json:"type" gorm:"column:type"` // 任务类型
	Status        enum.JobStatusEnum // 当前状态
	StartPosition string             // 任务起点
	EndPosition   string             // 任务终点
	StartUserId   string             // 开启任务的用户
	EndUserId     string             // 结束任务的用户
	StartTime     string             // 开始时间
	ArrivalTime   string             // 到达时间
	EndTime       string             // 结束时间
	Back          int
	Process       string // 规划路线
	GroupId       string
	Origin        enum.MsgOriginEnum
	BaseModel
}

type RobotJobVo struct {
	RobotId         string
	MsgId           string
	JobId           string
	GroupId         string
	JobStartSpotId  string // 任务起点位置id(任务可能会被调度系统拆分了多个任务的，因此这里为当前任务位置起始id，而startSpotId始终未原始任务的起始位置id)
	StartSpotId     string // 起始位置id(原始任务位置id)
	EndSpotId       string // 目标位置id
	EndSpotHalt     string // 目标位置楼宇id(发送给机器人的)
	EndSpotName     string // 目标位置名称
	EndSpotFullName string // 目标位置全名
	/**
	 * 任务类型
	 * 10-配送任务
	 * 20-呼叫任务
	 * 30-返程任务（去系统默认原点 position type=7）
	 * 40-等待任务
	 * 50-充电任务（去充电点 position type=4）
	 */
	JobType              enum.JobTypeEnum
	UserId               string    // 发起任务的用户id
	Back                 int       // 是否立即返程
	Process              []string  // 规划路线id
	Floor                string    // 楼层
	AreaId               string    // 位置所在的区域id
	JobOrder             int       // 任务排序
	LiftSpotId           string    // 电梯口位置id
	LiftId               string    // 电梯id
	StartTime            string    // 开始时间
	FinalJobId           string    // 最终的任务id
	FinalJobType         int       // 最终的任务的类型
	FinalEndSpotId       string    // 最终的任务目标位置id
	FinalEndSpotName     string    // 最终的任务目标位置名称
	FinalEndSpotFullName string    // 最终的任务目标位置全名
	FinalFloor           string    // 最终的任务的目标楼层
	FinalEndSpotHalt     string    // 最终的目标位置楼宇id(发送给机器人的)
	ErrorRetryCount      int       // 因为错误而重试的次数
	CreateDate           time.Time `json:"-"` // 创建时间
	GroupJobId           string    // 当前执行的任务组中的任务id(当前真正执行的任务id)
	Remarks              string    // 任务来源
}

type JobData struct {
	// 任务
	Job RobotJobVo `json:"job"`
	// 任务调度的时候最原始的任务(最终需要调度的任务，因为充电调度的时候，充电位置(job)与oriJob可能是不一样的)
	OriJob RobotJobVo `json:"oriJob"`
	// 是否已经释放了的任务（结束）
	Release bool `json:"release"`
	// 结束的时间戳
	EndTime int64 `json:"endTime"`
	// 结束的状态(什么状态导致该任务结束的)
	EndStatus enum.RobotStatusEnum `json:"endStatus"`
	// 乘坐电梯完成后（不管有没有乘坐电梯），最终的任务id（如果是跨楼宇的任务，那么这里的值为跨楼宇等待点的值）；
	// 如果需要找到最终任务，那么查找RobotJobVo中的final相关的字段
	EndLiftFinalJobId string `json:"endLiftFinalJobId"`
	// 乘坐电梯完成后（不管有没有乘坐电梯），最终的任务类型（如果是跨楼宇的任务，那么这里的值为跨楼宇等待点的值）；
	// 如果需要找到最终任务，那么查找RobotJobVo中的final相关的字段
	EndLiftFinalJobType int `json:"endLiftFinalJobType"`
	// 乘坐电梯完成后（不管有没有乘坐电梯），最终的任务位置id（如果是跨楼宇的任务，那么这里的值为跨楼宇等待点的值）；
	// 如果需要找到最终任务，那么查找RobotJobVo中的final相关的字段
	EndLiftFinalEndSpotId string `json:"endLiftFinalEndSpotId"`
}

func (RobotJobLog) TableName() string {
	return "device_robot_job_log"
}

func (e *RobotJobLog) BeforeCreate(_ *gorm.DB) (err error) {
	e.Id = util.CreateUUID()
	e.CreateDate = time.Now()
	return
}

type RobotJobLog struct {
	Id         string
	RobotId    string // 机器人任务日志父类
	LogType    int    `json:"type" gorm:"column:type"` // 任务类型
	Status     int    // 状态
	PositionId string // 位置id
	JobId      string
	CreateDate time.Time
}

type RobotJobCancelVo struct {
	JobId  string
	UserId string
}
