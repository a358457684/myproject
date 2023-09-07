package handler

import (
	"common/biz/manager"
	"common/orm"
	"errors"
)

type mapEntity struct {
	RobotModel manager.RobotType //机器人底盘型号
	Resolution float64           //分辨率（米/像素）
	Width      int               //地图宽度
	Height     int               //地图高度
	OriginX    float64           //左下角坐标X
	OriginY    float64           //坐下角坐标Y
}

func (m mapEntity) TableName() string {
	return "device_robot_floor_map"
}

//位置坐标转换的配置
type coordinateEntity struct {
	DatumRobotModel       manager.RobotType //基准的机器人类型，默认显示MIR
	MappingRobotModel     manager.RobotType //映射的机器人类型，默认显示云际
	MappingWidthZoomRate  float64           //映射机器人宽的缩放比
	MappingHeightZoomRate float64           //映射机器人高的缩放比
	DatumMapHeight        int               //映射MIR地图高（填写可以降低偏差）
	DatumMapWidht         int               //映射MIR地图宽（填写可以降低偏差）
	MappingMapRotate      float64           //映射地图需要旋转角度
	DeviationX            float64           //x轴偏差值（填写可以减少偏差）
	DeviationY            float64           //y轴偏差值（填写可以减少偏差）
}

func (coordinateEntity) TableName() string {
	return "device_robot_coordinate_config"
}

func getMap(buildingID string, floor int, robotModel manager.RobotType) (mapEntity, error) {
	var result mapEntity
	if err := orm.DB.First(&result, "building_id=? and floor=? and robot_model=? and del_flag=0", buildingID, floor, robotModel).Error; err != nil {
		return result, err
	}
	return result, nil
}

func getCoordinate(buildingId string, floor int, type1, type2 manager.RobotType) (coordinateEntity, error) {
	var result coordinateEntity
	if err := orm.DB.First(&result,
		"building_id=? and floor=? and (datum_robot_model=? and mapping_robot_model=? or datum_robot_model=? and mapping_robot_model=? ) and del_flag=0",
		buildingId, floor, type1, type2, type2, type1).Error; err != nil {
		return result, err
	}
	return result, nil
}

// ConvertPoint 将srcType的point转换成desType的point.
func ConvertPoint(buildingId string, floor int, srcType, desType manager.RobotType, point Point) (Point, error) {
	if srcType == desType {
		return point, nil
	}
	srcMap, err := getMap(buildingId, floor, srcType)
	if err != nil {
		return Point{}, errors.New("查找原始地图信息失败")
	}
	desMap, err := getMap(buildingId, floor, desType)
	if err != nil {
		return Point{}, errors.New("查找目标地图信息失败")
	}
	coordinateEntity, err := getCoordinate(buildingId, floor, desType, srcType)
	if err != nil {
		return Point{}, errors.New("地图映射关系查找失败")
	}
	srcMapInfo := BaseMapInfo{
		RobotType:  srcMap.RobotModel,
		Resolution: srcMap.Resolution,
		Width:      srcMap.Width,
		Height:     srcMap.Height,
		RealOrigin: Point{
			X: srcMap.OriginX,
			Y: srcMap.OriginY,
		},
	}
	desMapInfo := BaseMapInfo{
		RobotType:  desMap.RobotModel,
		Resolution: desMap.Resolution,
		Width:      desMap.Width,
		Height:     desMap.Height,
		RealOrigin: Point{
			X: desMap.OriginX,
			Y: desMap.OriginY,
		},
	}
	convertCfg := MapPoint2OtherMapPointCfg{
		BaseRobotType:         coordinateEntity.DatumRobotModel,
		MappingRobotType:      coordinateEntity.MappingRobotModel,
		BaseMapHeight:         coordinateEntity.DatumMapHeight,
		BaseMapWidth:          coordinateEntity.DatumMapWidht,
		MappingHeightZoomRate: coordinateEntity.MappingHeightZoomRate,
		MappingWidthZoomRate:  coordinateEntity.MappingWidthZoomRate,
		MappingRotate:         coordinateEntity.MappingMapRotate,
		OffsetX:               coordinateEntity.DeviationX,
		OffsetY:               coordinateEntity.DeviationY,
	}
	desPoint := convertCfg.ConvertRealPoint(point, &srcMapInfo, &desMapInfo)
	return desPoint, nil
}

//判断point是否在areaPoints形成的区域之内
func IsPointInArea(buildingId string, floor int, pointType, areaType manager.RobotType, point Point, areaPoints ...Point) (bool, error) {
	if len(areaPoints) < 3 {
		return false, errors.New("区域至少需要三个点位信息")
	}
	point, err := ConvertPoint(buildingId, floor, pointType, areaType, point)
	if err != nil {
		return false, err
	}
	areaPointPtrArray := make([]Point, len(areaPoints))
	for _, areaPoint := range areaPoints {
		areaPointPtrArray = append(areaPointPtrArray, areaPoint)
	}
	return NewPolygon(areaPointPtrArray).PtInPolygon(&point), nil
}
