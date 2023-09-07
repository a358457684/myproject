package api

import (
	"eps_common/huping/net_resuse"
	"epshealth-airobot-monitor/controller"
	_ "epshealth-airobot-monitor/docs"
	monitorWebsocket "epshealth-airobot-monitor/monitor_websocket"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"micro-common1/config"
	"micro-common1/log"
	"net/http"
	"runtime/debug"
)

func Init() {

	router := gin.New()
	// 开启跨域
	router.Use(cors())

	// monitor_websocket
	router.GET("/api/websocket/robots", monitorWebsocket.WsPage)

	// 配置拦截器
	router.Use(logMiddleware, gin.Recovery(), errMiddleware)

	if config.Data.Project.Swagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Infof("swagger: http://localhost:%d/swagger/index.html", config.Data.Project.Port)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	apiRouter := router.Group("/api")

	// 获取token
	apiRouter.POST("/login", controller.Login)

	apiRouter.Use(jwtAuthMiddleware)

	// 获取权限菜单
	apiRouter.GET("/auth/user/permissions", controller.FindPermissions)
	// 获取机器人列表
	apiRouter.POST("/robotUser/robotList", controller.FindRobotList)
	// 获取省市
	apiRouter.GET("/robotUser/areaList", controller.FindAreaList)
	// 获取机构列表
	apiRouter.POST("/robotUser/office/list", controller.FindOffices)
	// 获取机器人配置
	apiRouter.GET("/robot/getRobotConfig", controller.GetRobotConfig)
	// 获取楼宇列表
	apiRouter.POST("/robotUser/getOfficeBuildingList", controller.GetOfficeBuilding)
	// 获取地图监控的楼层信息
	apiRouter.POST("/robotUser/getFloorList", controller.GetFloorList)
	// 获取监控地图信息
	apiRouter.POST("/robotUser/floorMap", controller.FloorMapAndRobot)
	// 查看机器人的状态列表信息
	apiRouter.POST("/device/getRobotStatus", controller.GetRobotStatus)
	// 查看机器人状态的源数据
	apiRouter.GET("/device/getSourceRobotStatus/:documentId", controller.GetSourceRobotStatus)
	// 查看机器人的任务列表信息
	apiRouter.POST("/robotJobQueue/list", controller.RobotJobQueueList)
	// 查看机器人的任务详情
	apiRouter.POST("/robotJobQueue/jobRecordList", controller.JobRecordList)
	// 获取推送列表
	apiRouter.POST("/robotPushMessage/list", controller.RobotPushMessageList)
	// 获取代理服务
	apiRouter.GET("proxyServer/list", controller.FindProxyServer)

	// 以下需要权限验证和保存日志
	apiRouter.Use(operationInterceptor)

	// 操作权限
	// 取消机器人任务
	apiRouter.POST("/robotJobQueue/cancelRobotJob", controller.CancelRobotJob)
	// 移除机器人
	apiRouter.GET("/robotUser/remove/:officeId/:robotId", controller.RemoveRobot)
	// 操作机器人
	apiRouter.POST("/robotUser/operateRobot", controller.OperateRobot)

	// 调度权限
	// 释放所有资源、取消所有任务、移除缓存任务
	apiRouter.POST("/robotJob/dispatchOperate", controller.DispatchOperate)

	listener, err := net_resuse.Listen("tcp", fmt.Sprintf(":%d", config.Data.Project.Port))
	if err == nil {
		err = router.RunListener(listener)
	}
	if err != nil {
		log.WithError(err).Info("http服务启动失败")
		panic(err)
	}
}

// 开启跨域函数
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Origin")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("Panic info is: %v", err)
				log.Errorf("Panic info is: %s", debug.Stack())
			}
		}()
		c.Next()
	}
}
