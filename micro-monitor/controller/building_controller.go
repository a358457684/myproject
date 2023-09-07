package controller

import (
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/result"
	"github.com/gin-gonic/gin"
)

// 机构
type OfficeData struct {
	OfficeId string `json:"id" binding:"required"` // 编号
}

// @Tags buildingAndFloorInfo
// @Summary 获取楼宇列表
// @Description 获取楼宇列表
// @Security ApiKeyAuth
// @Param param body OfficeData true "机构id"
// @Success 200 {object} result.Result{data=[]dao.OfficeBuildingVo}
// @Router /robotUser/getOfficeBuildingList [post]
func GetOfficeBuilding(c *gin.Context) {
	var officeVo OfficeData
	err := c.ShouldBind(&officeVo)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	officeBuildings := dao.GetBuildingByOfficeId(officeVo.OfficeId)
	result.Success(c, officeBuildings)
}
