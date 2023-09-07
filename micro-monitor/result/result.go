package result

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type BackType int

const (
	Succeed        BackType = 0   // 成功
	Failed         BackType = 1   // 失败
	LoginFail      BackType = 101 // 登录失败（帐号不存在、用户被锁定、密码错误、用户机构已注销）
	NoPermission   BackType = 102 // 无访问权限
	NoConfig       BackType = 103 // 没有找到配置
	ParamError     BackType = 400 // 参数错误
	InvalidToken   BackType = 401 // token无效
	WsInvalidToken BackType = 402 // ws的token无效
)

const (
	NoPermissionMsg = "用户没有权限使用该功能！"
)

func (tp BackType) Code() int {
	return int(tp)
}

func (tp BackType) String() string {
	return strconv.Itoa(tp.Code())
}

type Result struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Custom(c *gin.Context, tp BackType, message string) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    tp.String(),
		"message": message,
	})
	if tp != Succeed {
		_ = c.Error(errors.New(message))
	}
}

func BadRequest(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    ParamError.String(),
		"message": "参数错误",
	})
	_ = c.Error(err)
}

func Fail(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    Failed.String(),
		"message": message,
	})
	_ = c.Error(errors.New(message))
}

func Success(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code":    Succeed.String(),
		"message": "success",
		"data":    data,
	})
}
