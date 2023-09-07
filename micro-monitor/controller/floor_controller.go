package controller

import (
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/result"
	"github.com/gin-gonic/gin"
	"micro-common1/biz/manager"
)

type RobotFloorVo struct {
	SimpleRobotVo
	BuildingId string            `json:"buildingId"` // 楼宇id
	RobotModel manager.RobotType `json:"robotModel"` // 机器人类型
}

// @Tags buildingAndFloorInfo
// @Summary 获取地图监控的楼层信息
// @Description 获取地图监控的楼层信息
// @Security ApiKeyAuth
// @Param param body RobotFloorVo true "请求信息"
// @Success 200 {object} result.Result{data=[]string}
// @Router /robotUser/getFloorList [post]
func GetFloorList(c *gin.Context) {
	var vo RobotFloorVo
	err := c.ShouldBind(&vo)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	if vo.RobotModel == "" {
		vo.RobotModel = dao.GetByRobotId(vo.RobotId).Model
	}
	floors := dao.FindFloorByCondition(dao.FloorMapVo{
		OfficeId:         vo.OfficeId,
		RobotModel:       vo.RobotModel,
		OfficeBuildingId: vo.BuildingId,
	})
	result.Success(c, floors)
}
