package model

import (
	"micro-common1/biz/enum"
	"micro-common1/biz/manager"
	"time"
)

type RobotStatusRes struct {
	RobotId    string               `json:"robotId"`
	RobotName  string               `json:"robotName"`
	RobotModel manager.RobotType    `json:"robotModel,omitempty"` // 机器人型号
	Status     enum.RobotStatusEnum `json:"status"`
	X          float64              `json:"x"`
	Y          float64              `json:"y"`
}

// 返回给web端vo
type RobotStatusVo struct {
	OfficeId            string               `json:"officeId"`   // 机构
	OfficeName          string               `json:"officeName"` // 机构名称
	RobotId             string               `json:"robotId"`
	Name                string               `json:"name"`       // 机器人名称
	RobotModel          manager.RobotType    `json:"robotModel"` // 机器人型号
	RobotAccount        string               `json:"robotAccount"`
	Electric            float64              `json:"electric"` // 电量
	Floor               int                  `json:"floor"`    // 楼层
	BuildingId          string               `json:"buildingId"`
	BuildingName        string               `json:"buildingName"`
	X                   float64              `json:"x"`
	Y                   float64              `json:"y"`
	LastUploadTime      time.Time            `json:"lastUploadTime"`      // 最后上传时间
	Status              enum.RobotStatusEnum `json:"status"`              // 状态
	StatusText          string               `json:"statusText"`          // 状态描述
	NetStatus           int                  `json:"netStatus"`           // 网络状态
	NetStatusText       string               `json:"netStatusText"`       // 网络状态描述
	ChassisSerialNumber string               `json:"chassisSerialNumber"` // 软件版本号
	SoftVersion         string               `json:"softVersion"`         // 底盘版本号
	DispatchMode        bool                 `json:"dispatchMode"`        // 是否调度模式
	PauseType           int                  `json:"pauseType"`           // 是否暂停  0：正常， 1：暂停
	EStopStatus         int                  `json:"eStopStatus"`         // 是否急停  0：正常， 1：急停
}

type RobotStatusUpload struct {
	DocumentId      string
	RobotId         string
	RobotName       string
	RobotAccount    string
	OfficeId        string
	OfficeName      string
	RobotModel      string // 机器人型号
	BuildingName    string
	Status          enum.RobotStatusEnum
	StatusText      string
	JobId           string
	GroupId         string
	LastUploadTime  int64
	X               float64
	Y               float64
	Z               float64
	Orientation     float64
	SpotId          string // 机器人当前位置（就是最后一次到达的位置）
	SpotName        string
	Target          string // 目标位置
	TargetName      string
	Process         []string // 线路规则中要经过的位置
	NextSpot        string   // 下一个位置
	Floor           int      // 楼层
	Electric        float64  // 电量
	Message         string
	StartIdleTime   time.Time // 开始空闲的时间
	StartErrorTime  time.Time // 开始异常时间
	NetStatus       enum.NetStatusEnum
	JobType         int
	PauseType       int                // 是否暂停：1：暂停，0：正常
	EStopStatus     int                // 是否急停 0-正常，1-急停状态
	BuildingId      string             // 楼宇ID
	StatusStartTime int64              // 状态开始时间
	StatusEndTime   int64              // 状态结束时间
	TimeConsume     float64            // 时长
	ExecStateEnum   enum.ExecStateEnum // 任务执行状态
	FinalJobId      string
	DispatchMode    int
	FinalJobType    int
}
