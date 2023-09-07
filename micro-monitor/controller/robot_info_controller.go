package controller

import (
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/service"
	"epshealth-airobot-monitor/utils"
	"github.com/gin-gonic/gin"
	"micro-common1/biz/cache"
	"micro-common1/biz/enum"
	"time"
)

type RobotByOfficeAndStatusVo struct {
	OfficeId  string               `json:"officeId" binding:"required"`
	Status    enum.RobotStatusEnum `json:"status"`
	NetStatus netStatus            `json:"netStatus"`
}

type RobotConfigVo struct {
	StatusList    []enum.SimpleRobotStatus `json:"statusList"`
	JobStatusList []enum.SimpleJobStatus   `json:"jobStatusList"`
	JobTypeList   []enum.SimpleJobType     `json:"jobTypeList"`
	Models        []string                 `json:"models"`
}

type netStatus int

const (
	all        netStatus = iota + 3 // 全部
	connect                         // 已连接(仅表示redis中存有机器人状态)
	notConnect                      // 未连接
)

// @Tags auth
// @Summary 获取用户对应的机器人列表信息
// @Description 获取用户对应的机器人列表信息
// @Security ApiKeyAuth
// @Param param body RobotByOfficeAndStatusVo true "请求信息"
// @Success 200 {object} result.Result{data=[]model.RobotStatusVo}
// @Router /robotUser/robotList [post]
func FindRobotList(c *gin.Context) {

	// 获取当前登录用户的信息
	user := utils.GetJwtData(c)

	var vo RobotByOfficeAndStatusVo
	err := c.ShouldBind(&vo)
	if err != nil {
		result.BadRequest(c, err)
		return
	}

	robots := dao.FindRobotByUser(vo.OfficeId, user)
	configs := dao.FindAllOfficeConfigMode()

	var robotInfos []model.RobotStatusVo
	// 离线的机器人
	var offLineRobots []model.RobotStatusVo
	for _, robot := range robots {
		// 获取机器人状态
		status, err := cache.GetRobotStatus(robot.OfficeId, robot.Id)
		if err != nil {
			if vo.NetStatus == all && vo.Status == 0 {
				offLineRobots = append(offLineRobots, service.ToWebOffLineRobotInfo(robot, configs, int(notConnect)))
			}
			continue
		}
		if (vo.Status != 0 && vo.Status != status.RobotStatus) || !canAddRobot(status.NetStatus, vo.NetStatus) {
			continue
		}
		statusVo := service.ToWebRobotStatus(robot, configs, status)
		robotInfos = append(robotInfos, statusVo)
	}
	robotInfos = append(robotInfos, offLineRobots...)
	result.Success(c, robotInfos)
}

func canAddRobot(netStatus enum.NetStatusEnum, selectCode netStatus) bool {
	return selectCode == all || selectCode == connect || int(selectCode) == netStatus.Code()
}

// @Tags robot
// @Summary 查看机器人的状态列表信息
// @Description 查看机器人的状态列表信息
// @Security ApiKeyAuth
// @Param param body service.ElasticRobotStatusPageQuery true "请求信息"
// @Success 200 {object} result.Result{data=model.PageResult}
// @Router /device/getRobotStatus [post]
func GetRobotStatus(c *gin.Context) {
	var elasticRobotStatusPageQuery service.ElasticRobotStatusPageQuery
	err := c.ShouldBind(&elasticRobotStatusPageQuery)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	page := service.FindPageRobotStatus(c, elasticRobotStatusPageQuery)
	result.Success(c, page)
}

// @Tags robot
// @Summary 查看机器人的任务列表信息
// @Description 查看机器人的任务列表信息
// @Security ApiKeyAuth
// @Param param body model.RobotJobQueryVo true "请求信息"
// @Success 200 {object} result.Result{data=model.PageResult}
// @Router /robotJobQueue/list [post]
func RobotJobQueueList(c *gin.Context) {
	var vo model.RobotJobQueryVo
	err := c.ShouldBind(&vo)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	page := service.FindPageRobotJob(vo)
	result.Success(c, page)
}

// @Tags robot
// @Summary 获取推送列表
// @Description 获取推送列表
// @Security ApiKeyAuth
// @Param param body service.ElasticRobotPushMessagePageQuery true "请求信息"
// @Success 200 {object} result.Result{data=model.PageResult}
// @Router /robotPushMessage/list [post]
func RobotPushMessageList(c *gin.Context) {
	var elasticRobotPushMessagePageQuery service.ElasticRobotPushMessagePageQuery
	err := c.ShouldBind(&elasticRobotPushMessagePageQuery)
	if err != nil {
		result.BadRequest(c, err)
		return
	}
	page := service.FindPageRobotPushMessage(c, elasticRobotPushMessagePageQuery)
	result.Success(c, page)
}

// @Tags robot
// @Summary 获取任务执行记录
// @Description 获取任务执行记录
// @Security ApiKeyAuth
// @Param param body model.RobotJobStatusChangeQuery true "获取任务记录数据VO"
// @Success 200 {object} result.Result{data=[]model.ElasticRobotJobExec}
// @Router /robotJobQueue/jobRecordList [post]
func JobRecordList(c *gin.Context) {
	var query model.RobotJobStatusChangeQuery
	if err := c.ShouldBind(&query); err != nil {
		result.BadRequest(c, err)
		return
	}
	robotJob := dao.GetRobotJobById(query.JobId)
	// 设置date,用于设置使用的索引
	if robotJob.Id != "" {
		query.Day = robotJob.CreateDate
	} else {
		query.Day = time.Now()
	}
	robotStatuses := service.FindRobotJobExecRecordList(c, query)
	result.Success(c, robotStatuses)
}

// @Tags robot
// @Summary 获取所有机器人相关配置
// @Description 获取所有机器人相关配置(机器人类型、机器人状态)
// @Security ApiKeyAuth
// @Success 200 {object} result.Result{data=RobotConfigVo}
// @Router /robot/getRobotConfig [get]
func GetRobotConfig(c *gin.Context) {
	status := enum.GetAllRobotStatus()
	jobStatus := enum.GetAllJobStatus()
	jobType := enum.GetAllJobType()
	models := dao.FindRobotModels()
	result.Success(c, RobotConfigVo{status, jobStatus, jobType, models})
}

// @Tags robot
// @Summary 获取机器人状态的源数据
// @Description 获取机器人状态的源数据
// @Security ApiKeyAuth
// @Param documentId path string true "文档Id"
// @Success 200 {object} result.Result{data=object}
// @Router /device/getSourceRobotStatus/{documentId} [get]
func GetSourceRobotStatus(c *gin.Context) {
	service.GetSourceRobotStatus(c)
}
