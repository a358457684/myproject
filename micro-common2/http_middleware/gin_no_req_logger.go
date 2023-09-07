package http_middleware

import (
	"github.com/gin-gonic/gin"
)

const IsNoLogHTTPRequest = "isNoLogHttpRequest"

// 记录http响应结果
func NoReqLogger() gin.HandlerFunc {
	return func(ctxt *gin.Context) {
		ctxt.Set(IsNoLogHTTPRequest, true)

		ctxt.Next()
	}
}
