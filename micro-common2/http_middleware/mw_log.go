package http_middleware

import (
	"mime"
	"net/http"
	"time"

	"pp/common-golang/logger"

	"github.com/spf13/viper"
	"pp/common-golang/ginx"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger *logger.Logger, skippers ...SkipperFunc) gin.HandlerFunc {

	return func(c *gin.Context) {

		start := time.Now()
		//c.Set(HTTPRequestStartTime, start)
		//reqReader := &reqBodyLogReader{buffer: &bodyBuffer{Buffer: *bytes.NewBufferString("")}, ReadCloser: c.Request.Body}
		//c.Request.Body = reqReader
		//respWriter := &respBodyLogWriter{buffer: &bodyBuffer{Buffer: *bytes.NewBufferString("")}, ResponseWriter: c.Writer}
		//c.Writer = respWriter
		if SkipHandler(c, skippers...) {
			c.Next()
			return
		}

		method := c.Request.Method

		entry := logger.WithContext(c)

		fields := make(map[string]interface{})
		fields["ip"] = c.ClientIP()
		fields["method"] = method
		fields["url"] = c.Request.URL.String()
		fields["header"] = c.GetHeader("Authorization")
		fields["user_agent"] = c.GetHeader("User-Agent")

		if method == http.MethodPost || method == http.MethodPut {
			mediaType, _, _ := mime.ParseMediaType(c.GetHeader("Content-Type"))
			if mediaType != "multipart/form-data" {
				if v, ok := c.Get(ginx.ReqBodyKey); ok {
					if b, ok := v.([]byte); ok && len(b) <= viper.GetInt("http.max_logger_length") {
						fields["body"] = string(b)
					}
				}
			}
		}
		c.Next()
		if c.Writer.Status() < http.StatusBadRequest {

			timeConsuming := time.Since(start).Nanoseconds() / 1e6
			fields["status"] = c.Writer.Status()
			fields["length"] = c.Writer.Size()

			//if v, ok := c.Get(ginx.LoggerReqBodyKey); ok {
			//	if b, ok := v.([]byte); ok && len(b) <= viper.GetInt("http.max_logger_length"){
			//		fields["body"] = string(b)
			//	}
			//}

			if v, ok := c.Get(ginx.ResBodyKey); ok {
				if b, ok := v.([]byte); ok && len(b) <= viper.GetInt("http.max_logger_length") {
					fields["responseBody"] = string(b)
				}
			}

			fields["user_id"] = ginx.GetUserID(c)
			entry.WithFields(fields).Infof("[http] %s-%s-%d(%dms)", c.Request.Method, c.ClientIP(), c.Writer.Status(), timeConsuming)
		}
	}
}
