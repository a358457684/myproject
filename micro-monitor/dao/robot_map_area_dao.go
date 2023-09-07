package dao

import (
	"epshealth-airobot-monitor/model"
	"micro-common1/orm"
	"time"
)

func FindMapAreaList(officeId string, buildingId string, floor int) (entries []model.RobotMapArea) {
	orm.DB.Select("id").
		Find(&entries, "del_flag = '0' and office_id = ? and building_id = ? and floor = ?",
			officeId, buildingId, floor)
	return
}

func FindMapAreaListByOfficeIdAndPosition(officeId string, startPosition string, endPosition string) model.RobotMapArea {
	var mapArea model.RobotMapArea
	orm.DB.Select("id", "job_id", "area_job_id", "update_date").
		First(&mapArea, "del_flag = '0' and office_id = ? and start_position = ? and end_position = ?",
			officeId, startPosition, endPosition)
	return mapArea
}

func SaveAreaJobRelation(areaJobRelation model.AreaJobRelation) {
	orm.DB.Create(&areaJobRelation)
}

func SaveRobotJobArea(robotJobArea model.RobotJobArea) {
	orm.DB.Create(&robotJobArea)
}

func FindAreaJobRelationByAreaIdAndJobId(areaId string, jobId string) (area model.AreaJobRelation) {
	orm.DB.Select("id").First(&area, "del_flag = '0' and area_id = ? and job_id = ?", areaId, jobId)
	return
}

func GetAreaJobRelation(areaJobId string) (area model.AreaJobRelation) {
	orm.DB.First(&area, areaJobId)
	return
}

func UpdateAreaJobRelationEndTimeById(id string, endTime time.Time) int64 {
	res := orm.DB.Exec(`UPDATE device_area_job_relation 
								SET end_time = ?
								WHERE id = ?`, id, endTime)
	return res.RowsAffected
}

func UpdateAreaJobIdById(area model.RobotMapArea) int64 {
	res := orm.DB.Exec(`UPDATE device_robot_job_area 
								SET area_job_id = ? ,
								update_date =  ?  
								WHERE id = ?`, area.AreaJobId, area.UpdateDate, area.Id)
	return res.RowsAffected
}

// 管制区域
func FindTrafficAreaByOfficeInfo(vo model.OfficeFloorVo) (areas []string) {
	orm.DB.Table("device_control_traffic_area").Select("area_coord").
		Find(&areas, "del_flag = '0' and office_id = ? and floor = ? and robot_model = ? and building_id = ?",
			vo.OfficeId, vo.Floor, vo.RobotModel, vo.BuildingId)
	return
}

// 电梯区域
func FindLiftTrafficAreaByOfficeInfo(vo model.OfficeFloorVo) (areas []string) {
	orm.DB.Table("device_lift_control_traffic_area").Select("area_coord").
		Find(&areas, "del_flag = '0' and office_id = ? and floor = ? and robot_model = ? and building_id = ?",
			vo.OfficeId, vo.Floor, vo.RobotModel, vo.BuildingId)
	return
}
