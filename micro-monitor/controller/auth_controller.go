package controller

import (
	"encoding/base64"
	"epshealth-airobot-monitor/dao"
	"epshealth-airobot-monitor/result"
	"epshealth-airobot-monitor/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"micro-common1/config"
	"micro-common1/log"
)

type LoginVo struct {
	Username string `json:"username" binding:"required"` // 账号
	Password string `json:"password" binding:"required"` // 密码
}

// @Tags auth
// @Summary 监控系统登录接口
// @Description 监控系统登录接口
// @Param param body LoginVo true "账号信息"
// @Success 200 {object} result.Result{data=string}
// @Router /login [post]
func Login(c *gin.Context) {
	var loginVo LoginVo
	if err := c.ShouldBind(&loginVo); err != nil {
		result.BadRequest(c, err)
		return
	}
	user := dao.GetUser(loginVo.Username)
	if user.Id == "" {
		result.Custom(c, result.LoginFail, "账号不存在")
		return
	}
	// 效验用户是否被限制登陆 0不可用 1可用
	if user.LoginFlag != "1" {
		result.Custom(c, result.LoginFail, "用户已被锁定,请联系管理员！")
		return
	}
	if !validatePassWd(user.Password, loginVo.Password) {
		result.Custom(c, result.LoginFail, "用户名或密码错误")
		return
	}
	data := utils.JwtData{
		Id:           user.Id,
		Username:     user.LoginName,
		HasAllOffice: user.HasAllOffice,
	}
	err := checkOffice(data)
	if err != nil {
		result.Custom(c, result.LoginFail, "用户机构已注销，无法登陆")
		return
	}
	jwtOptions := config.Data.Jwt
	if jwtOptions == nil {
		result.Custom(c, result.NoConfig, "JWT配置加载失败")
		return
	}
	token, err := utils.GenToken(data, "user_dao", jwtOptions.Issuer, jwtOptions.Secret, jwtOptions.Expires)
	if err != nil {
		log.WithError(err).Error("生成Token失败")
		result.Fail(c, "生成Token失败")
		return
	}
	result.Success(c, token)
}

// 比对用户密码是否相等
func validatePassWd(src string, password string) bool {
	passWd, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(src), passWd)
	return err == nil
}

// 如果是全机构不校验
func checkOffice(user utils.JwtData) error {
	if !user.HasAllOffice {
		offices := dao.FindOffices(user)
		if len(offices) == 0 {
			return errors.New("用户机构已注销，无法登陆")
		}
	}
	return nil
}

// @Tags auth
// @Summary 获取权限菜单
// @Description 获取用户的权限菜单
// @Security ApiKeyAuth
// @Success 200 {object} result.Result{data=[]dao.MenuVo}
// @Router /auth/user/permissions [get]
func FindPermissions(c *gin.Context) {
	user := utils.GetJwtData(c)
	if dao.IsAdmin(user) {
		permissions := dao.FindAllPermissions()
		result.Success(c, permissions)
		return
	}
	// 查询相应的权限菜单
	permissions := dao.FindPermissionsByUserId(user.Id)
	result.Success(c, permissions)
}
