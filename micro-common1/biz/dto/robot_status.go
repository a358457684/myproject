package dto

import (
	"common/biz/enum"
	"common/biz/manager"
	"time"
)

// 机器人状态
type RobotStatus struct {
	OfficeId           string               // 机构id
	OfficeName         string               // 机构名称
	BuildingId         string               `json:",omitempty"` // 楼宇ID
	BuildingName       string               `json:",omitempty"` // 楼宇名称
	RobotId            string               // 机器人id
	RobotName          string               // 机器人名称
	RobotModel         manager.RobotType    // 机器人型号
	RobotStatus        enum.RobotStatusEnum // 状态
	Electric           float64              // 电量
	Floor              int                  // 楼层
	Pause              bool                 // 是否暂停
	EStop              bool                 // 是否急停
	Orientation        float64              // 角度
	X                  float64              // 坐标x
	Y                  float64              // 坐标y
	LastPositionId     string               `json:",omitempty"` // 上次位置点id
	LastPositionName   string               `json:",omitempty"` // 上次位置点名称
	NextPositionId     string               `json:",omitempty"` // 下一次位置点id
	NextPositionName   string               `json:",omitempty"` // 下一次位置点名称
	TargetPositionId   string               `json:",omitempty"` // 最终目标位置点id
	TargetPositionName string               `json:",omitempty"` // 最终目标位置点名称
	GroupId            string               `json:",omitempty"` // 任务组id
	JobId              string               `json:",omitempty"` // 任务id
	JobType            enum.JobTypeEnum     `json:",omitempty"` // 任务类型
	Time               time.Time            // 时间
	NetStatus          enum.NetStatusEnum   `json:",omitempty"` // 网络状态
	StartIdleTime      time.Time            // 开始空闲时间（老版）
	StartErrorTime     time.Time            // 开始故障时间（老版）
	EstopStatus        int                  // 是否急停  0-正常，1-急停状态（老版）
	PauseType          int                  // 暂停状态（老版）
	Process            []string             // 途经位置（老版）
}

type RobotStatusToPAD struct {
	RobotID            string               `json:"robotID"`                      //机器人ID
	RobotName          string               `json:"robotName"`                    //机器人名称
	RobotModel         manager.RobotType    `json:"robotModel"`                   //机器人类型
	BuildingName       string               `json:"buildingName"`                 // 楼宇名称
	Status             enum.RobotStatusEnum `json:"status"`                       // 状态
	StatusDescription  string               `json:"statusDescription,omitempty"`  //状态描述
	LastPositionId     string               `json:"lastPositionId,omitempty"`     //上一个位置点
	LastPositionName   string               `json:"lastPositionName,omitempty"`   //上一个位置点名称
	TargetPositionId   string               `json:"targetPositionId,omitempty"`   //目标位置点
	TargetPositionName string               `json:"targetPositionName,omitempty"` //目标位置点名称
	JobId              string               `json:"jobId,omitempty"`              //当前任务ID
	JobGroup           string               `json:"jobGroup,omitempty"`           //当前任务组
	JobType            enum.JobTypeEnum     `json:",omitempty"`                   // 当前任务类型
	Floor              int                  `json:"floor"`                        // 楼层
	Electric           float64              `json:"electric"`                     // 电量
	OnLine             bool                 `json:"onLine"`                       //网路状态
	Pause              bool                 `json:"pause"`                        //是否暂停
	EStop              bool                 `json:"estop"`                        //是否急停
	X                  float64              `json:"x"`                            //当前坐标位置
	Y                  float64              `json:"y"`                            //当前坐标位置
}

func NewRobotStatusToPAD(status RobotStatus) RobotStatusToPAD {
	return RobotStatusToPAD{
		RobotID:            status.RobotId,
		RobotName:          status.RobotName,
		RobotModel:         status.RobotModel,
		BuildingName:       status.BuildingName,
		Status:             status.RobotStatus,
		StatusDescription:  status.RobotStatus.Description(),
		LastPositionName:   status.LastPositionName,
		LastPositionId:     status.LastPositionId,
		TargetPositionId:   status.TargetPositionId,
		TargetPositionName: status.TargetPositionName,
		JobId:              status.JobId,
		JobGroup:           status.GroupId,
		JobType:            status.JobType,
		Floor:              status.Floor,
		Electric:           status.Electric,
		OnLine:             status.NetStatus == enum.NsOnline,
		Pause:              status.Pause,
		EStop:              status.EStop,
		X:                  status.X,
		Y:                  status.Y,
	}
}
