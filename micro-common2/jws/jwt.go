package jws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pp/common-golang/ginx"
	"pp/common-golang/storer"
	"pp/common-golang/utils"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// JWT 签名结构
type JWT struct {
	SigningKey []byte
	Storer     storer.Storer
}

// 一些常量
var (
	TokenExpired     = errors.New("授权已过期")
	TokenNotValidYet = errors.New("授权未生效")
	TokenMalformed   = errors.New("无权限访问")
	TokenInvalid     = errors.New("无权限访问")
	ServerErr        = errors.New("服务错误")
	TokenHeaderName  = "Authorization"
	Claims           = "claims"
	DefaultSignKey   = "defaultSignKey"
	TokenTimeout     = 1209600 // 60 * 60 * 24 * 14
	UserRedisKey     = "token"
	UserInfoRedisKey = "login_user_info:"
)

// 载荷，可以加一些自己需要的信息
type TokenClaims struct {
	UserId   int64  `json:"userId"`
	TenantId int64  `json:"tenantId"`
	Token    string `json:"-"`
	jwt.StandardClaims
}

// 新建一个jwt实例
func NewJWT(storer storer.Storer, args ...string) *JWT {
	var signKey string
	if key := viper.GetString("jws.sign_key"); len(strings.TrimSpace(key)) > 0 {
		signKey = strings.TrimSpace(key)
	} else if len(args) > 0 {
		signKey = strings.TrimSpace(args[0])
	} else {
		signKey = DefaultSignKey
	}
	return &JWT{Storer: storer, SigningKey: []byte(signKey)}
}

func (j *JWT) GetToken(ctxt *gin.Context) string {
	if nil == ctxt || nil == ctxt.Request ||
		nil == ctxt.Request.Header || len(ctxt.Request.Header) == 0 {
		return ""
	}
	token := ctxt.Request.Header.Get(TokenHeaderName)
	if token == "" {
		token = ctxt.Request.Header.Get(strings.ToLower(TokenHeaderName))
	}

	return strings.TrimSpace(token)
}

// CreateToken 生成一个token
func (j *JWT) CreateToken(ctx context.Context, claims TokenClaims) (string, error) {
	jwt.TimeFunc = time.Now
	if claims.StandardClaims.ExpiresAt <= 0 {
		claims.StandardClaims.ExpiresAt = time.Now().Add(time.Duration(TokenTimeout) * time.Second).Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["exp"] = claims.StandardClaims.ExpiresAt
	tokenString, err := token.SignedString(j.SigningKey)
	if err != nil {
		return "", err
	}

	userIdKey := fmt.Sprintf("%s:%d", UserRedisKey, claims.UserId)
	_ = j.Storer.DelToken(ctx, userIdKey)
	if err := j.Storer.SetToken(ctx, userIdKey, tokenString, time.Duration(TokenTimeout)*time.Second); err != nil {
		return "", err
	}

	return tokenString, nil
}

// 解析Tokne
func (j *JWT) ParseToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, TokenInvalid
		}

		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				panic(utils.E2(http.StatusProxyAuthRequired, 102, TokenMalformed.Error(), ""))
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				panic(utils.E2(http.StatusProxyAuthRequired, 102, TokenExpired.Error(), ""))
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				panic(utils.E2(http.StatusProxyAuthRequired, 102, TokenNotValidYet.Error(), ""))
			} else {
				panic(utils.E2(http.StatusProxyAuthRequired, 102, TokenInvalid.Error(), ""))
			}
		}
	}
	if nil == token {
		return nil, TokenInvalid
	}
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {

		userIdKey := fmt.Sprintf("%s:%d", UserRedisKey, claims.UserId)
		redisToken, err := j.Storer.GetToken(ctx, userIdKey)
		if err != nil {
			return nil, ServerErr
		}
		if redisToken != tokenString {
			return nil, TokenInvalid
		}
		return claims, nil
	}
	return nil, TokenInvalid
}

// 更新token
func (j *JWT) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, TokenInvalid
		}

		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return j.CreateToken(ctx, *claims)
	}
	return "", TokenInvalid
}

// 将组织架构信息放入上下文中
func (j *JWT) CreateGroupInfo(ctxt *gin.Context) {
	userInfoKey := fmt.Sprintf("%s%s", UserInfoRedisKey, ginx.GetUserID(ctxt))
	data, err := j.Storer.GetStr(ctxt, userInfoKey)
	if err != nil {
		return
	}
	var userInfoMap map[string]interface{}
	_ = json.Unmarshal([]byte(data), &userInfoMap)
	if v, ok := userInfoMap["loginType"]; ok {
		ginx.SetLoginType(ctxt, v.(float64))
	}

	if v, ok := userInfoMap["storeId"]; ok {
		ginx.SetStoreID(ctxt, v.(string))
	}

	if v, ok := userInfoMap["regionId"]; ok {
		ginx.SetRegionID(ctxt, v.(string))

	}
	if v, ok := userInfoMap["drivingSchoolId"]; ok {
		ginx.SetDrivingSchoolID(ctxt, v.(string))
	}
}
