package utils

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"micro-common1/config"
	"micro-common1/util"
	"strings"
	"time"
)

type JwtData struct {
	Id           string
	Username     string
	HasAllOffice bool
}

type MyClaims struct {
	JwtData
	jwt.StandardClaims
}

func GenToken(data JwtData, audience, issuer, secret string, expires time.Duration) (string, error) {
	c := MyClaims{
		JwtData: data,
		StandardClaims: jwt.StandardClaims{
			Id:        util.CreateUUID(),                // id
			Subject:   fmt.Sprintf("%sToken", audience), // 主题
			Audience:  audience,                         // 接收方
			Issuer:    issuer,                           // 签发者
			NotBefore: time.Now().Unix(),                // 生效时间
			IssuedAt:  time.Now().Unix(),                // 签发时间
			ExpiresAt: time.Now().Add(expires).Unix(),   // 过期时间
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	return fmt.Sprintf("bearer %s", token), err
}

func ParseToken(tokenString string) (JwtData, error) {
	if index := strings.Index(tokenString, " "); index >= 0 {
		tokenString = tokenString[index+1:]
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Data.Jwt.Secret), nil
	})
	if err != nil {
		return JwtData{}, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims.JwtData, nil
	}
	return JwtData{}, errors.New("invalid token")
}

func GetJwtData(c *gin.Context) JwtData {
	return c.MustGet("jwtData").(JwtData)
}
