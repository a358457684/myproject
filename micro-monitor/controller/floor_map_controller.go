package controller

import (
	"context"
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/result"
	"fmt"
	"github.com/gin-gonic/gin"
	"micro-common1/biz/cache"
	"micro-common1/biz/handler"
	"micro-common1/log"
	"micro-common1/redis"
	"time"
)

type FloorMapAndRobotVo struct {
	OfficeId          string                 `json:"officeId"`
	MapFile           string                 `json:"mapFile"`
	Floor             int                    `json:"floor"`
	OriginX           float64                `json:"originX"`
	OriginY           float64                `json:"originY"`
	Resolution        float64                `json:"resolution"`
	PositionList      []dao.RobotPositionRes `json:"positionList"`      // 充电桩集合
	RobotStatusList   []model.RobotStatusRes `json:"robotStatusList"`   // 机器人状态集合
	TrafficAreaCoords []model.Point          `json:"trafficAreaCoords"` // 管制区域坐标集合
	LiftAreaCoords    []model.Point          `json:"liftAreaCoords"`    // 电梯区域坐标集合
}

// @Tags map
// @Summary 获取监控地图信息
// @Description 获取监控地图信息
// @Security ApiKeyAuth
// @Param param body model.OfficeFloorVo true "请求信息"
// @Success 200 {object} result.Result{data=FloorMapAndRobotVo}
// @Router /robotUser/floorMap [post]
func FloorMapAndRobot(c *gin.Context) {
	var vo model.OfficeFloorVo
	err := c.ShouldBind(&vo)
	if err != nil {
		result.BadRequest(c, err)
		return
	}

	// 取地图
	robotMap := dao.FindMapByCondition(vo)
	if robotMap.Id == "" {
		result.Fail(c, "该机构没有找到地图信息，请在管理后台核查！")
		return
	}

	floorMapAndRobotResult := FloorMapAndRobotVo{
		Floor:      vo.Floor,
		OfficeId:   vo.OfficeId,
		OriginX:    robotMap.OriginX,
		OriginY:    robotMap.OriginY,
		Resolution: robotMap.Resolution,
	}

	mapInfo := handler.BaseMapInfo{
		RobotType:  vo.RobotModel,
		Resolution: robotMap.Resolution,
		Width:      robotMap.Width,
		Height:     robotMap.Height,
		RealOrigin: handler.Point{
			X: robotMap.OriginX,
			Y: robotMap.OriginY,
		},
	}

	redisKey := fmt.Sprintf("%s:%s:%d", constant.MonitorMap, vo.BuildingId, vo.Floor)
	_ = redis.SetJson(context.Background(), redisKey, mapInfo, time.Hour*3)

	if robotMap.FreehandMapFile == "" {
		floorMapAndRobotResult.MapFile = robotMap.MapFile
	} else {
		floorMapAndRobotResult.MapFile = robotMap.FreehandMapFile
	}

	// 取点位
	positionList := dao.FindRobotPositionByCondition(vo)
	var positionResultList []dao.RobotPositionRes
	for _, position := range positionList {
		robotPositionRes := dao.RobotPositionRes{
			Id:           position.GuId,
			Cname:        position.Name,
			PositionType: position.PositionType,
		}
		point := handler.Point{X: position.X, Y: position.Y}
		point = mapInfo.PixelPointByRealPoint(point, float64(robotMap.Height))
		robotPositionRes.Cx = point.X
		robotPositionRes.Cy = point.Y
		positionResultList = append(positionResultList, robotPositionRes)
	}
	floorMapAndRobotResult.PositionList = positionResultList

	// 取当前机构、楼层的机器人
	robots, _ := cache.FindRobotStatusByOfficeId(vo.OfficeId)
	var robotStatusResultList []model.RobotStatusRes
	for _, robotStatus := range robots {
		if robotStatus.Floor == vo.Floor {
			robotStatusRes := model.RobotStatusRes{
				RobotId:   robotStatus.RobotId,
				RobotName: robotStatus.RobotName,
				Status:    robotStatus.RobotStatus,
			}
			robotModel := robotStatus.RobotModel.Type()
			checkModel := vo.RobotModel.Type()
			if robotModel == nil || checkModel == nil || robotModel.Chassis == nil || checkModel.Chassis == nil {
				continue
			}
			point := handler.Point{X: robotStatus.X, Y: robotStatus.Y}
			if robotModel.Chassis.Supplier != checkModel.Chassis.Supplier {
				point, err = handler.ConvertPoint(vo.BuildingId, vo.Floor, robotStatus.RobotModel, vo.RobotModel, point)
				log.WithError(err).Errorf("多机器地图坐标转换失败")
			}
			point = mapInfo.PixelPointByRealPoint(point, float64(robotMap.Height))
			robotStatusRes.X = point.X
			robotStatusRes.Y = point.Y
			robotStatusResultList = append(robotStatusResultList, robotStatusRes)
		}
	}
	floorMapAndRobotResult.RobotStatusList = robotStatusResultList

	trafficAreaList := dao.FindTrafficAreaByOfficeInfo(vo)
	floorMapAndRobotResult.TrafficAreaCoords = getPixelPoint(trafficAreaList, mapInfo)

	liftAreaList := dao.FindLiftTrafficAreaByOfficeInfo(vo)
	floorMapAndRobotResult.LiftAreaCoords = getPixelPoint(liftAreaList, mapInfo)

	result.Success(c, floorMapAndRobotResult)
}

func getPixelPoint(areaStrList []string, mapInfo handler.BaseMapInfo) []model.Point {
	for _, areaStr := range areaStrList {
		areas := make([]model.Point, 0)
		_ = json.Unmarshal([]byte(areaStr), &areas)
		if len(areas) != 4 {
			continue
		}
		for index, coordinate := range areas {
			point := handler.Point{X: coordinate.X, Y: coordinate.Y}
			point = mapInfo.PixelPointByRealPoint(point, float64(mapInfo.Height))
			areas[index] = model.Point{X: point.X, Y: point.Y}
		}
		return areas
	}
	return make([]model.Point, 0)
}
