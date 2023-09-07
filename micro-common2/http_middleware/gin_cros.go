package http_middleware

import "github.com/gin-gonic/gin"

// 允许跨域响应头中间件
func AddCrossOriginHeaders() gin.HandlerFunc {
	return func(ctxt *gin.Context) {
		ctxt.Header("Access-Control-Allow-Credentials", "true")
		var origin string
		if ctxt.Request.Header.Get("Origin") != "" {
			origin = ctxt.Request.Header.Get("Origin")
		} else {
			origin = "*"
		}
		ctxt.Header("Access-Control-Allow-Origin", origin)
		ctxt.Header("Access-Control-Max-Age", "3600")
		ctxt.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE,PUT")
		ctxt.Header("Access-Control-Allow-Headers", "Origin,x-requested-with,If-Modified-Since,Pragma,Last-Modified,Cache-Control,Expires,Content-Type,X-E4M-With,Accept,Authorization,Platform")
		ctxt.Next()
	}
}
