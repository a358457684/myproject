package controller

import (
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/utils"
	"github.com/gin-gonic/gin"
)

// @Tags auth
// @Summary 获取省市区域列表
// @Description 获取省市区域列表
// @Security ApiKeyAuth
// @Success 200 {object} result.Result{data=[]dao.AreaVo}
// @Router /robotUser/areaList [get]
func FindAreaList(c *gin.Context) {
	result.Success(c, dao.FindAreaList())
}

// @Tags auth
// @Summary 获取用户的机构列表
// @Description 获取用户的机构列表
// @Security ApiKeyAuth
// @Success 200 {object} result.Result{data=[]dao.OfficeVo}
// @Router /robotUser/office/list [post]
func FindOffices(c *gin.Context) {
	user := utils.GetJwtData(c)
	offices := dao.FindOffices(user)
	result.Success(c, offices)
}
