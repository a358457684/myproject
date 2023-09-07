package api

import (
	"encoding/json"
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/controller"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"micro-common1/log"
	"micro-common1/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func logMiddleware(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	start := time.Now()             // 开始时间
	c.Next()                        // 执行
	end := time.Now()               // 结束时间
	latency := end.Sub(start)       // 运行时间
	path := c.Request.URL.Path      // 访问地址
	clientIp := c.ClientIP()        // 客户端ip
	method := c.Request.Method      // 请求方式
	statusCode := c.Writer.Status() // 响应状态
	var level logrus.Level
	if statusCode < http.StatusBadRequest {
		level = logrus.InfoLevel
	} else if statusCode < http.StatusInternalServerError {
		level = logrus.WarnLevel
	} else {
		level = logrus.ErrorLevel
	}
	fields := logrus.Fields{
		"method":     method,
		"clientIp":   clientIp,
		"statusCode": statusCode,
		"latency":    latency,
	}
	log.WithFields(fields).Log(level, path)
}

func errMiddleware(c *gin.Context) {
	c.Next()
	status := c.Writer.Status()
	errors := c.Errors
	if len(errors) > 0 {
		log.Errorf(fmt.Sprintf("请求:%s, 错误信息:%d -> %s", c.Request.RequestURI, status, errors.String()))
	}
	if status == http.StatusOK {
		return
	}
	switch status {
	case http.StatusBadRequest:
		c.JSON(status, result.Result{Code: strconv.Itoa(status), Message: "参数错误"})
	case http.StatusNotFound:
		c.JSON(status, result.Result{Code: strconv.Itoa(status), Message: "未知请求"})
	case http.StatusUnauthorized, http.StatusNotAcceptable, http.StatusInternalServerError:
		c.JSON(status, result.Result{Code: strconv.Itoa(status), Message: errors.String()})
	}
}

func jwtAuthMiddleware(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		result.Custom(c, result.InvalidToken, "访问未授权")
		return
	}
	jwtData, err := utils.ParseToken(token)
	if err != nil {
		result.Custom(c, result.InvalidToken, "访问未授权")
		return
	}
	c.Set("jwtData", jwtData)
	c.Next()
}

// 权限日志拦截器
func operationInterceptor(c *gin.Context) {
	user := utils.GetJwtData(c)
	// 无法获取登录信息
	if user.Id == "" || user.Username == "" {
		// 无权限日志
		saveLog(c, user.Id, dao.TypeException, result.NoPermissionMsg)
		result.Custom(c, result.NoPermission, result.NoPermissionMsg)
		return
	}
	// 如果不是管理员
	if !dao.IsAdmin(user) {
		// 权限认证
		permission := constant.GetPermission(c.Request.RequestURI)
		menuVo := dao.FindMenuByUser(permission, user.Id)
		if menuVo.Id == "" {
			// 无权限日志
			saveLog(c, user.Id, dao.TypeException, result.NoPermissionMsg)
			result.Custom(c, result.NoPermission, result.NoPermissionMsg)
			return
		}
	}
	c.Next()
	if c.Errors == nil {
		// 正常日志
		saveLog(c, user.Id, dao.TypeAccess, "")
	} else {
		// 异常
		saveLog(c, user.Id, dao.TypeException, c.Errors.String())
	}
}

func saveLog(c *gin.Context, userId, logType, errorMessage string) {
	log.Infof("===========用户:%s 操作日志，类型:%s============", userId, logType)
	uri := c.Request.RequestURI
	operateLog := dao.Log{
		Id:         util.CreateUUID(),
		LogType:    logType,
		RemoteAddr: c.Request.RemoteAddr,
		UserAgent:  c.Request.UserAgent(),
		RequestUri: uri,
		Method:     c.Request.Method,
		Title:      getTitle(c, uri),
		Exception:  errorMessage,
		SystemType: dao.MonitorSystem,
		CreateBy:   userId,
		CreateDate: time.Now().Format(constant.DateTimeFormat),
	}
	query, err := json.Marshal(c.Request.PostForm)
	if err != nil {
		operateLog.Params = string(query)
	}
	_ = dao.InsertLog(operateLog)
}

func getTitle(c *gin.Context, uri string) string {
	// 操作机器人（重启底盘、退出APP、重启APP、恢复急停、上传日志）
	if constant.PtOperateRobot.Code() == uri {
		var vo controller.OperateRobotVo
		_ = c.ShouldBindBodyWith(&vo, binding.JSON)
		return fmt.Sprintf("监控系统-%s", vo.OperateType)
	}
	// 调度操作（释放所有资源、取消所有任务、移除缓存任务）
	if constant.PtDispatchOperate.Code() == uri {
		var vo controller.ReleaseVo
		_ = c.ShouldBindBodyWith(&vo, binding.JSON)
		return fmt.Sprintf("监控系统-%s", vo.ReleaseType)
	}
	// 移除机器人
	if strings.HasPrefix(uri, constant.PtRemoveRobot.Code()) {
		return constant.PtRemoveRobot.String()
	}
	return constant.PermissionType(uri).String()
}
