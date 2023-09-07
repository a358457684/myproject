package model

import (
	"epshealth-airobot-monitor/constant"
	"micro-common1/biz/enum"
	"micro-common1/biz/manager"
	"time"
)

type ElasticRobotStatus struct {
	DocumentId     string               `json:"documentId"` // 文档id
	RobotId        string               `json:"robotId"`
	OfficeId       string               `json:"officeId"`
	RobotModel     manager.RobotType    `json:"robotModel"` // 机器人型号
	BuildingName   string               `json:"buildingName"`
	BuildingId     string               `json:"buildingId"` // 楼宇ID
	Status         enum.RobotStatusEnum `json:"status"`
	StatusText     string               `json:"statusText"`
	JobId          string               `json:"jobId,omitempty"`
	LastUploadTime time.Time            `json:"lastUploadTime"`
	X              float64              `json:"x"`
	Y              float64              `json:"y"`
	SpotId         string               `json:"spotId,omitempty"` // 最后一个位置
	SpotName       string               `json:"spotName,omitempty"`
	Target         string               `json:"target,omitempty"` // 目标位置
	TargetName     string               `json:"targetName,omitempty"`
	NextSpot       string               `json:"nextSpot,omitempty"` // 下一个位置
	Floor          int                  `json:"floor"`              // 楼层
	Electric       float64              `json:"electric"`           // 电量
	NetStatus      enum.NetStatusEnum   `json:"netStatus"`
	NetStatusText  string               `json:"netStatusText"`
	PauseType      int                  `json:"pauseType"`   // 是否暂停  0：正常， 1：暂停
	EStopStatus    int                  `json:"eStopStatus"` // 是否急停  0：正常， 1：急停
}

type ElasticRobotPushMessage struct {
	DocumentId     string              `json:"documentId"` // 文档id
	RobotId        string              `json:"robotId,omitempty"`
	OfficeId       string              `json:"officeId,omitempty"`
	Timestamp      time.Time           `json:"timestamp,omitempty"`      // 消息时间
	MsgId          string              `json:"msgId,omitempty"`          // 消息id
	Path           string              `json:"path,omitempty"`           // 消息路径
	Body           string              `json:"body,omitempty"`           // 消息内容
	Status         constant.PushStatus `json:"status"`                   // 推送状态
	StatusText     string              `json:"statusText"`               // 推送状态描述
	FirstTimestamp time.Time           `json:"firstTimestamp,omitempty"` // 第一次发送时的消息时间
	SendCount      int                 `json:"sendCount"`                // 发送的次数
}

type ElasticRobotJobExec struct {
	DocumentId   string               `json:"documentId"`             // 文档id
	RobotId      string               `json:"robotId,omitempty"`      // 机器人id
	RobotName    string               `json:"robotName,omitempty"`    // 机器人名称
	BuildingName string               `json:"buildingName,omitempty"` // 楼宇名称
	Status       enum.RobotStatusEnum `json:"status,omitempty"`
	StatusText   string               `json:"statusText,omitempty"`
	JobId        string               `json:"jobId,omitempty"`      // jobId
	SpotName     string               `json:"spotName,omitempty"`   // 当前位置名称
	TargetName   string               `json:"targetName,omitempty"` // 目标位置名称
	Floor        int                  `json:"floor,omitempty"`      // 楼层
	Message      string               `json:"message,omitempty"`
	PauseType    int                  `json:"pauseType,omitempty"`   // 是否暂停：1：暂停，0：正常
	EstopStatus  int                  `json:"estopStatus,omitempty"` // 是否急停 0-正常，1-急停状态
	// 新增加的字段
	ExecState       enum.ExecStateEnum    `json:"execStateEnum,omitempty"`   // 任务执行状态
	ExecStateText   string                `json:"execStateText,omitempty"`   // 任务执行状态描述
	StatusStartTime int64                 `json:"statusStartTime,omitempty"` // 状态开始时间
	StatusEndTime   int64                 `json:"statusEndTime,omitempty"`   // 状态结束时间
	TimeConsume     float64               `json:"timeConsume,omitempty"`     // 耗时
	StopInfo        string                `json:"stopInfo,omitempty"`
	FinalJobId      string                `json:"finalJobId,omitempty"`   // 最终任务Id
	DispatchMode    int                   `json:"dispatchMode,omitempty"` // 运行模式
	FinalJobType    string                `json:"finalJobType,omitempty"` // 最终任务类型
	AcceptState     enum.AcceptStatusEnum `json:"acceptState,omitempty"`  // 物品接收状态
}

// 查询任务记录
type RobotJobStatusChangeQuery struct {
	OfficeId string    `json:"officeId" binding:"required"`
	JobId    string    `json:"jobId" binding:"required"` // 任务id
	Day      time.Time `json:"day"`                      // 时间，主要用来定位索引
	RobotId  string    `json:"robotId"`                  // 机器人id
}

type ElasticSourceRobotStatus struct {
	DocumentId     string    `json:"documentId"`
	SourceMsg      string    `json:"sourceMsg"`
	RobotId        string    `json:"robotId"`
	OfficeId       string    `json:"officeId"`
	Status         string    `json:"status"`
	LastUploadTime time.Time `json:"lastUploadTime"`
}
