package dao

import (
	"epshealth-airobot-monitor/model"
	"micro-common1/biz/manager"
	"micro-common1/orm"
)

type FloorMapVo struct {
	Floor            int
	OfficeId         string
	RobotModel       manager.RobotType
	OfficeBuildingId string
}

type FloorMap struct {
	Id              string            // id
	MapFile         string            // 地图
	Floor           string            // 楼层
	RobotModel      manager.RobotType // 机器人地盘类型
	Resolution      float64           // 地图分辨率
	Width           int               // 地图宽度
	Height          int               // 地图高度
	OriginX         float64           // 左下角坐标X
	OriginY         float64           // 左下角坐标Y
	FreehandMapFile string            // 手绘地图
}

func (FloorMap) TableName() string {
	return "device_robot_floor_map"
}

func (FloorMapVo) TableName() string {
	return "device_robot_floor_map"
}

func FindFloorByCondition(vo FloorMapVo) (floors []string) {
	tx := orm.DB.
		Where("del_flag = '0'").
		Where("office_id = ?", vo.OfficeId).
		Where("robot_model = ?", vo.RobotModel).
		Where("building_id = ?", vo.OfficeBuildingId).
		Where("map_file IS NOT NULL")
	tx.Model(vo).Select("floor").Order("floor").Find(&floors)
	return
}

func FindMapByCondition(vo model.OfficeFloorVo) (floorMap FloorMap) {
	tx := orm.DB.
		Where("del_flag = '0'").
		Where("office_id = ?", vo.OfficeId).
		Where("floor = ?", vo.Floor).
		Where("robot_model = ?", vo.RobotModel).
		Where("building_id = ?", vo.BuildingId)
	tx.Select(`id, map_file, freehand_map_file, width, height, resolution, origin_x, origin_y`).Take(&floorMap)
	return
}
