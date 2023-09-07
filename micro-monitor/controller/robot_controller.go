package controller

import (
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/mq"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"micro-common1/biz/cache"
	"micro-common1/biz/dto"
	"micro-common1/biz/enum"
	bizMq "micro-common1/biz/mq"
	"micro-common1/biz/restful"
	"micro-common1/config"
	"micro-common1/log"
	"strings"
	"time"
)

type SimpleRobotVo struct {
	OfficeId string `json:"officeId" binding:"required"` // 机构id
	RobotId  string `json:"robotId" binding:"required"`  // 机器人id
}

type OperateRobotVo struct {
	SimpleRobotVo
	OperateType  constant.RobotOperationEnum `json:"operateType" binding:"required"`
	AlertMessage string                      `json:"alertMessage"` // 操作原因信息
}

type ReleaseVo struct {
	ReleaseType enum.DispatchReleaseEnum `json:"releaseType" binding:"required"`
	SimpleRobotVo
}

type ReleaseRobotFloorVo struct {
	SimpleRobotVo
	BuildingId string `json:"buildingId" binding:"required"` // 楼宇
	Floor      string `json:"floor" binding:"required"`      // 楼层
}

type CallRobotVo struct {
	EndSpotId string `json:"endSpotId" binding:"required"`
	Remarks   string `json:"remarks" binding:"required"`
	SimpleRobotVo
}

// @Tags robot
// @Summary 取消机器人任务
// @Description 取消机器人任务
// @Security ApiKeyAuth
// @Param param body SimpleRobotVo true "请求信息"
// @Success 200 {object} result.Result{}
// @Router /robotJobQueue/cancelRobotJob [post]
func CancelRobotJob(c *gin.Context) {

	// 获取当前登录用户的信息
	user := utils.GetJwtData(c)

	var vo SimpleRobotVo
	err := c.ShouldBind(&vo)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	robot, _ := cache.GetRobotStatus(vo.OfficeId, vo.RobotId)
	if robot.JobId == "" {
		result.Fail(c, "当前没有任务")
		return
	}
	if robot.RobotStatus.CanNotSendTask() {
		result.Fail(c, fmt.Sprintf("机器人在[%s]不允许取消任务", robot.RobotStatus.Description()))
		return
	}

	// 对数据进行封装，传值
	applyRobotJobVo := dto.RobotJobCompleted{
		BaseRobotJob: dto.BaseRobotJob{
			Origin:   enum.MoMonitor,
			OfficeID: vo.OfficeId,
			RobotID:  vo.RobotId,
			GroupID:  robot.GroupId,
			JobID:    robot.JobId,
		},
		Status:          enum.JsCancel,
		CompletedTime:   time.Now(),
		CompletedUserID: user.Id,
		Remarks:         enum.MoMonitor.String(),
	}
	log.Infof("取消机器人任务信息：%+v", applyRobotJobVo)
	err = restful.RobotJobCompleted(config.Data.RelatedServer.Dispatch, applyRobotJobVo)
	if err != nil {
		log.WithError(err).Errorf("取消机器人任务失败")
		result.Fail(c, "取消机器人任务失败")
		return
	}
	result.Custom(c, result.Succeed, "取消机器人任务成功")
}

// @Tags robot
// @Summary 操作机器人
// @Description 操作机器人
// @Security ApiKeyAuth
// @Param param body OperateRobotVo true "操作信息"
// @Success 200 {object} result.Result{}
// @Router /robotUser/operateRobot [post]
func OperateRobot(c *gin.Context) {
	var vo OperateRobotVo
	err := c.ShouldBindBodyWith(&vo, binding.JSON)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	robotStatus, err := cache.GetRobotStatus(vo.OfficeId, vo.RobotId)
	// 机器人离线不推送
	if err != nil || robotStatus.NetStatus == enum.NsOffline {
		result.Fail(c, "机器人离线状态，无法推送")
		return
	}
	if strings.Trim(vo.AlertMessage, " ") == "" {
		vo.AlertMessage = vo.OperateType.String()
	}
	cmdEnum := enum.RobotCmdEnum(vo.OperateType.Code())
	err = bizMq.SendRobotCmd(bizMq.NewRobotCmdDTO(vo.OfficeId, vo.RobotId, cmdEnum, nil, enum.MoMonitor, vo.AlertMessage))
	if err != nil {
		result.Fail(c, fmt.Sprintf("%s指令发送失败", vo.OperateType))
		return
	}
	result.Custom(c, result.Succeed, fmt.Sprintf("%s指令发送成功", vo.OperateType))
}

// @Tags robot
// @Summary 移除机器人
// @Description 移除机器人
// @Security ApiKeyAuth
// @Param officeId path string true "机构Id"
// @Param robotId path string true "机器人Id"
// @Success 200 {object} result.Result{}
// @Router /robotUser/remove/{officeId}/{robotId} [get]
func RemoveRobot(c *gin.Context) {
	officeId := c.Param("officeId")
	robotId := c.Param("robotId")
	_, err := cache.RemoveRobotStatus(officeId, robotId)
	if err == nil {
		result.Custom(c, result.Succeed, "移除机器人成功")
	}
	result.Fail(c, "移除机器人失败")
}

// @Tags dispatch
// @Summary 调度相关的操作
// @Description 调度相关的操作
// @Security ApiKeyAuth
// @Param param body ReleaseVo true "机构机器人信息"
// @Success 200 {object} result.Result{}
// @Router /robotJob/dispatchOperate [post]
func DispatchOperate(c *gin.Context) {
	var vo ReleaseVo
	err := c.ShouldBindBodyWith(&vo, binding.JSON)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	err = mq.PublishDispatch(vo.ReleaseType, vo.SimpleRobotVo)
	if err == nil {
		result.Custom(c, result.Succeed, fmt.Sprintf("%s指令发送成功", vo.ReleaseType))
		return
	}
	result.Fail(c, fmt.Sprintf("%s指令发送失败", vo.ReleaseType))
}
