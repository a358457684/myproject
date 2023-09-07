package http_middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"pp/common-golang/utils"
	"strconv"
	"time"
)

var (
	tsHeaderName   = "timestamp"
	signHeaderName = "sign"
)

func SignAuth(appId, appSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		timestamp := c.Request.Header.Get(tsHeaderName)
		timeInt, err := strconv.ParseInt(timestamp, 10, 64)
		if timestamp == "" || err != nil {
			res := utils.BanBanE(http.StatusUnauthorized, "无权限访问")
			c.JSON(http.StatusUnauthorized, res)
			c.Abort()
			return
		}

		if time.Now().Unix()*1000-(5*60*1000) > timeInt {
			res := utils.BanBanE(http.StatusUnauthorized, "签名已过期")
			c.JSON(http.StatusUnauthorized, res)
			c.Abort()
			return
		}
		headerSign := c.Request.Header.Get(signHeaderName)
		var bodyJson string
		if c.Request.Method == http.MethodPost {
			body, _ := ioutil.ReadAll(c.Request.Body)
			bodyJson = string(body)
			c.Request.Body.Close()
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
		}

		md := md5.New()
		md.Write([]byte(appId))
		md.Write([]byte(appSecret))
		md.Write([]byte(timestamp))
		md.Write([]byte(bodyJson))
		sign := hex.EncodeToString(md.Sum(nil))
		if sign == headerSign {
			c.Next()
		} else {
			res := utils.BanBanE(http.StatusUnauthorized, "无权限访问")
			c.JSON(http.StatusUnauthorized, res)
			c.Abort()
		}
	}
}
