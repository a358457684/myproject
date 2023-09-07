package model

import (
	"micro-common1/biz/enum"
	"micro-common1/biz/manager"
	"time"
)

type SimpleRobot struct {
	OfficeId string `json:"officeId" binding:"required"` // 机构id
	RobotId  string `json:"robotId" binding:"required"`  // 机器人id
}

type OfficeFloorVo struct {
	OfficeId   string            `json:"officeId" binding:"required"`   // 机构id
	BuildingId string            `json:"buildingId" binding:"required"` // 楼宇id
	Floor      int               `json:"floor" binding:"required"`      // 楼层
	RobotModel manager.RobotType `json:"robotModel" binding:"required"` // 机器人类型
}

// 分页时间查询
type BasePageQuery struct {
	OfficeId  string    `json:"officeId" binding:"required"`  // 机构id
	RobotId   string    `json:"robotId" binding:"required"`   // 机器人id
	PageIndex int       `json:"pageIndex" binding:"required"` // 页码
	PageSize  int       `json:"pageSize" binding:"required"`  // 页数
	StartDate time.Time `json:"startDate"`                    // 开始时间
	EndDate   time.Time `json:"endDate"`                      // 结束时间
}

// 分页数据
type PageResult struct {
	Total     int64       `json:"total"`
	Data      interface{} `json:"data"`
	PageIndex int         `json:"pageIndex"`
	PageSize  int         `json:"pageSize"`
}

// 任务详情查询
type RobotJobQueryVo struct {
	BasePageQuery
	JobType   enum.JobTypeEnum   `json:"jobType"`
	JobStatus enum.JobStatusEnum `json:"jobStatus"`
}
