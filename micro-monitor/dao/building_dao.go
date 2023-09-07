package dao

import "micro-common1/orm"

type OfficeBuildingVo struct {
	Id   string `json:"id"`   // 编号
	Name string `json:"name"` // 名称
}

func (OfficeBuildingVo) TableName() string {
	return "robot_building"
}

func GetBuildingByOfficeId(officeId string) (officeBuildings []OfficeBuildingVo) {
	orm.DB.Select("id", "name").
		Find(&officeBuildings, "del_flag = '0' and office_id = ?", officeId)
	return
}
