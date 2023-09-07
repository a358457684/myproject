package http_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"pp/common-golang/ginx"
	"pp/common-golang/logger"
	"pp/common-golang/utils"
	"strings"
)

// Recovery中间件（统一错误处理）
func Recovery(hostPrefix string, logger *logger.Logger) gin.HandlerFunc {
	return func(ctxt *gin.Context) {
		defer func() {
			if err := recover(); err != nil && logger != nil {
				stack := string(utils.GetStack(4))

				clientIP := ctxt.ClientIP()
				comment := ctxt.Errors.ByType(gin.ErrorTypePrivate).String()
				req, fields := splitUri(hostPrefix, ctxt)

				if v, ok := ctxt.Get(ginx.ReqBodyKey); ok {
					if b, ok := v.([]byte); ok && len(b) <= viper.GetInt("http.max_logger_length") {
						fields["body"] = string(b)
					}
				}

				logger = logger.
					WithField("clientIP", clientIP).
					WithField("comment", comment).
					WithFields(fields)

				if res, ok := err.(*utils.Res); ok {
					httpStatus := http.StatusOK
					var s1, s2 string
					if nil != &res.ResHead {
						s1 = res.Msg
						s2 = res.Detail
						if strings.Contains(s2, s1) {
							s1 = ""
						}
						if strings.Contains(s1, s2) {
							s2 = ""
						}
					} else {
						res.ResHead = utils.ResHead{Code: -1, Msg: "未定义"}
					}

					logger.
						WithCaller(4).
						WithField("tag", "Custom warn").
						WithField("httpStatus", httpStatus).
						WithField("errCode", res.Code).
						WithField("errMsg", strings.TrimSpace(fmt.Sprintf("%d %s %s", res.Code, s1, s2))).
						Warn(req)
					ctxt.JSON(httpStatus, res)
					ctxt.Abort()
				} else {
					res := utils.E2(-1, 500, "未定义", nil)
					//if i, ok := ctxt.Get(IsLogHTTPResponse); ok {
					//	if isLogHttpResponse, ok := i.(bool); ok && isLogHttpResponse {
					//		logger = logger.WithField("responseBody", `{"status":-1,"msg":"未定义"}`)
					//	}
					//}
					logger.
						WithCaller(5).
						WithField("tag", "Catch Exception").
						WithField("httpStatus", http.StatusBadRequest).
						WithField("errCode", -1).
						WithField("errMsg", fmt.Sprintf("-1 %s", err)).
						Error(req)
					println(stack)
					ctxt.JSON(http.StatusBadRequest, res)
					ctxt.Abort()
				}
			}
		}()
		ctxt.Next()
	}
}
