package handler

import (
	"common/biz/manager"
	"fmt"
	"testing"
)

func TestBaseMapInfo_RealRobotPositionByPixel(t *testing.T) {
	chasis := manager.Chassis{
		Supplier:          "mir",
		Name:              "mir",
		DefaultResolution: 0.05,
	}
	chasis.BindRobotType("E2", "测试E2", 0)
	manager.RegisterChassis(chasis)

	chasis = manager.Chassis{
		Supplier:          "云迹",
		Name:              "Y2",
		DefaultResolution: 0.03,
	}
	chasis.BindRobotType("Y2", "测试Y2", 0)
	manager.RegisterChassis(chasis)

	e2map := BaseMapInfo{
		RobotType:  "E2",
		Resolution: 0.05,
	}
	Y2Map := BaseMapInfo{
		RobotType:  "Y2",
		Resolution: 0.03,
		Width:      1389,
		Height:     1231,
		RealOrigin: Point{X: -21.0204, Y: -15.3298},
	}
	mapconfig := MapPoint2OtherMapPointCfg{
		BaseRobotType:         "E2",
		MappingRobotType:      "Y2",
		MappingHeightZoomRate: 59.10000,
		MappingWidthZoomRate:  60.60000,
		BaseMapHeight:         783,
		BaseMapWidth:          886,
		MappingRotate:         357.70,
		OffsetX:               31.26,
		OffsetY:               -5.96,
	}
	//本配置允许转换才转换
	if mapconfig.CanConvertPoint(Y2Map.RobotType, e2map.RobotType) {
		preal := Point{-0.68, 20.33}
		p := mapconfig.ConvertRealPoint(preal, &Y2Map, &e2map)
		fmt.Println("Y2地图下的真实环境坐标", preal, "转换到E2环境下的真实坐标为", p)
		preal = p
		p = mapconfig.ConvertRealPoint(preal, &e2map, &Y2Map)
		fmt.Println("E2地图下的真实环境坐标", preal, "转换到Y2环境下的真实坐标为", p)
	} else {
		fmt.Sprintf("配置无法实现从%s到%s的坐标转换", Y2Map.RobotType, e2map.RobotType)
	}
}
