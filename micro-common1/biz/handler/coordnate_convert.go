package handler

import (
	"common/biz/manager"
	"math"
)

//屏幕地图图片信息
type BaseMapInfo struct {
	RobotType  manager.RobotType //是哪个底盘的地图
	Resolution float64           //分辨率（米/像素）
	Width      int               //地图图片的宽度
	Height     int               //地图图片的高度
	RealOrigin Point             //地图图片的左下角在真实物里环境下的坐标信息，主要用来作为计算相对于坐标原点的地图偏移
}

/*****机器人在地图上的像素坐标转换到真实环境坐标公式
xreal = pixelX * Resolution + RealOrigin.X
因为屏幕上显示的图片坐标原点为左上角，需要转换到左下角坐标原点，所以需要地图高度减去
yreal = (Height - pixelY) * Resolution + RealOrigin.Y
*/
func (mapInfo *BaseMapInfo) RealRobotPositionByPixel(pixel Point, MapHeight float64) Point {
	if mapInfo.Resolution <= 0 {
		tp := mapInfo.RobotType.Type()
		if tp != nil {
			mapInfo.Resolution = tp.Chassis.DefaultResolution
			mapInfo.RealOrigin.X = tp.Chassis.DefaultRealOriginX
			mapInfo.RealOrigin.Y = tp.Chassis.DefaultRealOriginY
		}
	}

	var result Point
	result.X = pixel.X*mapInfo.Resolution + mapInfo.RealOrigin.X
	if MapHeight <= 0 {
		MapHeight = float64(mapInfo.Height)
	}
	if MapHeight > 0 {
		result.Y = (MapHeight-pixel.Y)*mapInfo.Resolution + mapInfo.RealOrigin.Y
	} else {
		result.Y = pixel.Y*mapInfo.Resolution + mapInfo.RealOrigin.Y
	}
	return result
}

//将一个真实环境下的机器人坐标转换到地图上的屏幕像素坐标
func (mapInfo *BaseMapInfo) PixelPointByRealPoint(rp Point, MapHeight float64) Point {
	if mapInfo.Resolution <= 0 {
		tp := mapInfo.RobotType.Type()
		if tp != nil {
			mapInfo.Resolution = tp.Chassis.DefaultResolution
			mapInfo.RealOrigin.X = tp.Chassis.DefaultRealOriginX
			mapInfo.RealOrigin.Y = tp.Chassis.DefaultRealOriginY
		}
	}
	var result Point
	if mapInfo.Resolution > 0 { // 不作处理,除以0,x y 会变成无穷大,json转化失败,导致panic
		result.X = (rp.X - mapInfo.RealOrigin.X) / mapInfo.Resolution
	} else {
		result.X = rp.X - mapInfo.RealOrigin.X
	}
	if MapHeight <= 0 {
		MapHeight = float64(mapInfo.Height)
	}
	if MapHeight > 0 {
		//result.Y = MapHeight - (rp.Y-mapInfo.RealOrigin.Y)/mapInfo.Resolution
		result.Y = (rp.Y - mapInfo.RealOrigin.Y) / mapInfo.Resolution
	} else {
		if mapInfo.Resolution > 0 {
			result.Y = (rp.Y - mapInfo.RealOrigin.Y) / mapInfo.Resolution
		} else {
			result.Y = rp.Y - mapInfo.RealOrigin.Y
		}
	}
	return result
}

/**
 * 坐标原点变换为左下角的
 * @param posPicOrg 图片左下角位置
 * @return 转换比差
 */
func getDifference(posPicOrg Point) Point {
	posOrg := Point{X: 0, Y: 0} //坐标原点
	var returnPoint Point
	returnPoint.X = posOrg.X - posPicOrg.X
	returnPoint.Y = posOrg.Y - posPicOrg.Y
	return returnPoint
}

/**
 * 坐标转换成像素
 *
 * @param point      坐标点
 * @param mPerPixel 米每像素
 * @return 像素点
 */
func positionToPixel(paramPoint Point, mPerPixel float64) Point {
	var returnPoint Point
	returnPoint.X = paramPoint.X * mPerPixel
	returnPoint.Y = paramPoint.Y * mPerPixel
	return returnPoint
}

/***将一个地图的坐标点转换到另外一个地图上的坐标的配置信息
思路原理，是使用一张BaseRobotType的地图作为基准地图，然后导入一张MapingRobotType类型的地图
通过操作变换（缩放，旋转，平移）之后达到让两个地图的各个轮廓最大程度的重合，之后MapingRobotType类型的地图
所调整的参数，然后通过这个参数来进行坐标计算变换
可以实现互相转换
*/
type MapPoint2OtherMapPointCfg struct {
	BaseRobotType         manager.RobotType //基准的机器人类型，就是以这个机器底盘作为标准坐标系，将其他的坐标都转换到这个地图上的坐标
	MappingRobotType      manager.RobotType //将要映射坐标到基准类型地图上的底盘类型
	BaseMapHeight         int               //用来产生这个配置参数的基准类型地图的宽和高
	BaseMapWidth          int
	MappingHeightZoomRate float64 //将MapRobotType的地图通过转换之后，高度的缩放比例
	MappingWidthZoomRate  float64 //将MapRobotType的地图通过转换之后，高度的缩放比例
	MappingRotate         float64 //将MapRobotType的地图通过转换之后，旋转的角度
	OffsetX               float64 //将MapRobotType的地图通过转换之后，在X轴上的平移
	OffsetY               float64 //将MapRobotType的地图通过转换之后，在Y轴上的平移
}

//通过这个配置，是否可以实现from底盘到to底盘的坐标转换
func (convertCfg *MapPoint2OtherMapPointCfg) CanConvertPoint(from, to manager.RobotType) bool {
	return from == convertCfg.BaseRobotType && to == convertCfg.MappingRobotType ||
		to == convertCfg.BaseRobotType && from == convertCfg.MappingRobotType
}

//从源地图上的像素点坐标转换到目标地图上的像素点坐标
func (convertCfg *MapPoint2OtherMapPointCfg) ConvertPixelPoint(point Point, sourceFloorMap, targetFloorMap *BaseMapInfo) Point {
	if sourceFloorMap.RobotType == targetFloorMap.RobotType {
		wscale := float64(targetFloorMap.Width) / float64(sourceFloorMap.Width)
		hscale := float64(targetFloorMap.Height) / float64(sourceFloorMap.Height)
		return Point{point.X * wscale, point.Y * hscale}
	}
	if sourceFloorMap.RobotType == convertCfg.BaseRobotType && targetFloorMap.RobotType == convertCfg.MappingRobotType {
		//从基准的配置逆向转换到原，所以需要做逆转换
		//第一步先平移回去
		point.X = point.X - convertCfg.OffsetX
		point.Y = point.Y - convertCfg.OffsetY

		mappingHeightZoomRate := convertCfg.MappingHeightZoomRate / 100
		mappingWidthZoomRate := convertCfg.MappingWidthZoomRate / 100
		//旋转
		if convertCfg.MappingRotate != 0 {
			//根据配置执行坐标换算
			destW := float64(targetFloorMap.Width) * mappingWidthZoomRate
			destH := float64(targetFloorMap.Height) * mappingHeightZoomRate
			//获取中心点的坐标，然后反向旋转回去
			centerp := Point{destW / 2, destH / 2}
			θ := (360 - convertCfg.MappingRotate) * math.Pi / 180
			//根据中心点旋转,新坐标公式
			/*x= (x1 - x2)*cos(θ) - (y1 - y2)*sin(θ) + x2 ;
			y= (x1 - x2)*sin(θ) + (y1 - y2)*cos(θ) + y2*/
			cosθ := math.Cos(θ)
			sinθ := math.Sin(θ)
			newx := (point.X-centerp.X)*cosθ - (point.Y-centerp.Y)*sinθ + centerp.X
			newY := (point.X-centerp.X)*sinθ + (point.Y-centerp.Y)*cosθ + centerp.Y
			point.X = newx
			point.Y = newY
		}
		point.X = point.X / mappingWidthZoomRate
		point.Y = point.Y / mappingHeightZoomRate
	} else if sourceFloorMap.RobotType == convertCfg.MappingRobotType && targetFloorMap.RobotType == convertCfg.BaseRobotType {
		//转换到配比的基准
		mappingHeightZoomRate := convertCfg.MappingHeightZoomRate / 100
		mappingWidthZoomRate := convertCfg.MappingWidthZoomRate / 100
		//以地图左上角为坐标进行计算
		point.X = point.X * mappingWidthZoomRate
		point.Y = point.Y * mappingHeightZoomRate
		if convertCfg.MappingRotate != 0 {
			//根据配置执行坐标换算
			destW := float64(sourceFloorMap.Width) * mappingWidthZoomRate
			destH := float64(sourceFloorMap.Height) * mappingHeightZoomRate
			//获取中心点的坐标
			centerp := Point{destW / 2, destH / 2}
			θ := convertCfg.MappingRotate * math.Pi / 180
			//根据中心点旋转,新坐标公式
			/*x= (x1 - x2)*cos(θ) - (y1 - y2)*sin(θ) + x2 ;
			y= (x1 - x2)*sin(θ) + (y1 - y2)*cos(θ) + y2*/
			cosθ := math.Cos(θ)
			sinθ := math.Sin(θ)
			newx := (point.X-centerp.X)*cosθ - (point.Y-centerp.Y)*sinθ + centerp.X
			newY := (point.X-centerp.X)*sinθ + (point.Y-centerp.Y)*cosθ + centerp.Y
			point.X = newx
			point.Y = newY
		}
		//地图坐标平移
		point.X = point.X + convertCfg.OffsetX
		point.Y = point.Y + convertCfg.OffsetY
	}
	return point
}

//从源地图上的真实点转换到另一个地图上的真实点
func (convertCfg *MapPoint2OtherMapPointCfg) ConvertRealPoint(point Point, sourceFloorMap, targetFloorMap *BaseMapInfo) Point {
	//转换到目标的真实点
	if sourceFloorMap.RobotType == convertCfg.MappingRobotType && targetFloorMap.RobotType == convertCfg.BaseRobotType {
		point = sourceFloorMap.PixelPointByRealPoint(point, 0)
		point = convertCfg.ConvertPixelPoint(point, sourceFloorMap, targetFloorMap)
		rh := targetFloorMap.Height
		if rh == 0 {
			rh = convertCfg.BaseMapHeight
		}
		return targetFloorMap.RealRobotPositionByPixel(point, float64(rh))
	} else if sourceFloorMap.RobotType == convertCfg.BaseRobotType && targetFloorMap.RobotType == convertCfg.MappingRobotType {
		rh := float64(sourceFloorMap.Height)
		if rh == 0 {
			rh = float64(convertCfg.BaseMapHeight)
		}
		point = sourceFloorMap.PixelPointByRealPoint(point, rh)
		point = convertCfg.ConvertPixelPoint(point, sourceFloorMap, targetFloorMap)
		return targetFloorMap.RealRobotPositionByPixel(point, float64(targetFloorMap.Height))
	} else {
		point = sourceFloorMap.PixelPointByRealPoint(point, 0)
		point = convertCfg.ConvertPixelPoint(point, sourceFloorMap, targetFloorMap)
		return targetFloorMap.RealRobotPositionByPixel(point, 0)
	}
}
