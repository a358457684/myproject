package dao

import (
	"epshealth-airobot-monitor/model"
	"micro-common1/biz/enum"
	"micro-common1/orm"
)

type RobotPositionRes struct {
	Id           string                `json:"id"`    // guid
	Cname        string                `json:"cname"` // 位置名称
	Cx           float64               `json:"cx"`    // 位置 x 坐标
	Cy           float64               `json:"cy"`    // 位置 y 坐标
	PositionType enum.PositionTypeEnum `json:"type"`  // 任务类型 1 、4 、7 以外的类型不能呼叫
}

type RobotPosition struct {
	Id           string
	Name         string
	FullName     string
	PositionType enum.PositionTypeEnum `json:"type" gorm:"column:type"` // 位置类型 4-充电点 7-默认原点
	X            float64
	Y            float64
	Z            float64
	Floor        string // 楼层
	BuildingId   string
	OfficeId     string
	GuId         string
	TeleNumber   string
	W            float32
	RobotModel   string
	Sort         int
}

func (RobotPosition) TableName() string {
	return "device_robot_position"
}

func FindRobotPositionByCondition(vo model.OfficeFloorVo) (robotPositions []RobotPosition) {

	tx := orm.DB.
		Where("del_flag = '0'").
		Where("office_id = ?", vo.OfficeId).
		Where("floor = ?", vo.Floor).
		Where("building_id = ?", vo.BuildingId)

	if vo.RobotModel != "" {
		tx.Where("robot_model = ?", vo.RobotModel)
	}

	tx.Model(RobotPosition{}).Select(`id, gu_id, name, full_name, type, x, y`).Find(&robotPositions)
	return
}

func FindRobotPositionByOfficeId(officeId string) (robotPositions []RobotPosition) {
	orm.DB.Select("name", "full_name", "gu_id", "building_id").
		Find(&robotPositions, "del_flag = '0' and office_id = ?", officeId)
	return
}

func GetRobotPositionByGuId(officeId, guId string) (robotPosition RobotPosition) {
	orm.DB.Select("type").
		Take(&robotPosition, "del_flag = '0' and office_id = ? and gu_id = ?", officeId, guId)
	return
}
