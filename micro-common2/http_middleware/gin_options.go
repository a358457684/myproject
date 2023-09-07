package http_middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// HTTP OPTIONS、HEAD 方法直接响应成功中间件
func HandleOptionsMethod() gin.HandlerFunc {
	return func(ctxt *gin.Context) {
		if strings.ToUpper(ctxt.Request.Method) == http.MethodOptions || strings.ToUpper(ctxt.Request.Method) == http.MethodHead {
			ctxt.String(http.StatusOK, "")
			//ctxt.Header("Content-Type","text/html; charset=utf-8")
			ctxt.Abort()
		}
	}
}
