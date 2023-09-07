package http_middleware

import (
	"github.com/gin-gonic/gin"
)

const IsLogHTTPResponse = "isLogHttpResponse"

// 记录http响应结果
func RespLogger() gin.HandlerFunc {
	return func(ctxt *gin.Context) {
		ctxt.Set(IsLogHTTPResponse, true)

		ctxt.Next()
	}
}
