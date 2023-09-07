package model

import (
	"gorm.io/gorm"
	"micro-common1/biz/manager"
	"micro-common1/util"
	"time"
)

func (e *AreaJobRelation) BeforeCreate(_ *gorm.DB) (err error) {
	e.Id = util.CreateUUID()
	e.CreateDate = time.Now()
	e.UpdateDate = time.Now()
	e.DelFlag = DelFlagNormal // 代表正常数据
	return
}

func (e *AreaJobRelation) BeforeUpdate(_ *gorm.DB) (err error) {
	e.UpdateDate = time.Now()
	return
}

func (e *RobotJobArea) BeforeCreate(_ *gorm.DB) (err error) {
	e.Id = util.CreateUUID()
	e.CreateDate = time.Now()
	e.UpdateDate = time.Now()
	e.DelFlag = DelFlagNormal // 代表正常数据
	return
}

func (e *RobotJobArea) BeforeUpdate(_ *gorm.DB) (err error) {
	e.UpdateDate = time.Now()
	return
}

func (AreaJobRelation) TableName() string {
	return "device_area_job_relation"
}

func (RobotJobArea) TableName() string {
	return "device_robot_job_area"
}

func (RobotMapArea) TableName() string {
	return "device_robot_map_area"
}

type RobotMapArea struct {
	Id         string
	AreaJobId  string
	JobId      string
	OfficeId   string // 所属机构
	OfficeName string // 所属机构
	BuildingId string // 机构楼宇ID
	Floor      string
	AreaName   string            // 区域名称
	RobotModel manager.RobotType // 机器人类型
	AreaCoord  string            // 区域坐标信息，json数据格式
	Color      string            // 区域颜色（用于回显）
	BaseModel
}

type AreaJobRelation struct {
	Id            string
	OfficeId      string    // 所属机构ID
	OfficeName    string    // 所属机构名称
	StartPosition string    // 任务起点
	EndPosition   string    // 任务终点
	AreaId        string    // 区域ID
	FinalJobId    string    // 最终任务ID
	JobId         string    // 任务ID
	StartTime     time.Time // 起始时间
	EndTime       time.Time // 结束时间
	BaseModel
}

type RobotJobArea struct {
	BaseModel
	Id                string
	OfficeId          string // 所属机构ID
	OfficeName        string // 所属机构名称
	BuildingId        string // 机构楼宇ID
	Floor             int    // 楼层
	StartPosition     string // 任务起点
	StartPositionName string
	EndPosition       string // 任务终点
	EndPositionName   string
	FinalJobId        string // 最终任务ID
	JobId             string // 任务ID
	AreaJobId         string // 任务区域关联ID(主要用于处理进入区域的结束时间)
	RobotModel        string // 机器人类型
	RobotMapAreaList  []RobotMapArea
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
