package ginx

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// 定义上下文中的键
const (
	UserIDKey          = "user-id"
	ReqBodyKey         = "req-body"
	ResBodyKey         = "res-body"
	StoreIdKey         = "store-id"
	RegionIdKey        = "area-id"
	DrivingSchoolIdKey = "driving-school-id"
	LoginType          = "login-type"
)

// GetToken 获取用户令牌
func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}
	return token
}

// GetUserID 获取用户ID
func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

// SetUserID 设定用户ID
func SetUserID(c *gin.Context, userID string) {
	c.Set(UserIDKey, userID)
}

// GetLoginType 获取loginType
func GetLoginType(c *gin.Context) float64 {
	return c.GetFloat64(LoginType)
}

// SetLoginType 设置loginType
func SetLoginType(c *gin.Context, loginType float64) {
	c.Set(LoginType, loginType)
}

// GetStoreID 获取门店ID
func GetStoreID(c *gin.Context) string {
	return c.GetString(StoreIdKey)
}

// SetStoreID 设置门店ID
func SetStoreID(c *gin.Context, storeId string) {
	c.Set(StoreIdKey, storeId)
}

// GetRegionID 获取片区ID
func GetRegionID(c *gin.Context) string {
	return c.GetString(RegionIdKey)
}

// SetRegionID 设置片区ID
func SetRegionID(c *gin.Context, regionId string) {
	c.Set(RegionIdKey, regionId)
}

// GetDrivingSchoolID 获取驾校ID
func GetDrivingSchoolID(c *gin.Context) string {
	return c.GetString(DrivingSchoolIdKey)
}

// SetDrivingSchoolID 设置驾校ID
func SetDrivingSchoolID(c *gin.Context, drivingSchoolId string) {
	c.Set(DrivingSchoolIdKey, drivingSchoolId)
}

// SetUserID 设定用户ID
func SetResBody(c *gin.Context, body []byte) {
	c.Set(ResBodyKey, body)
}

// GetBody Get request body
func GetBody(c *gin.Context) []byte {
	if v, ok := c.Get(ReqBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}

// ParseJSON 解析请求JSON
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	return nil
}

// ParseQuery 解析Query参数
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return err
	}
	return nil
}

// Header 解析Header参数
func ParseHeader(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindHeader(obj); err != nil {
		return err
	}
	return nil
}

// ParseForm 解析Form请求
func ParseForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return err
	}
	return nil
}
