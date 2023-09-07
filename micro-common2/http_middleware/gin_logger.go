package http_middleware

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"pp/common-golang/date"
	"pp/common-golang/logger"

	"github.com/gin-gonic/gin"
)

type bodyBuffer struct {
	bytes.Buffer
}

const MaxLogMessageLength = 32766

func (b *bodyBuffer) String() string {
	str := b.Buffer.String()
	if len(str) > MaxLogMessageLength {
		str = str[:MaxLogMessageLength]
	}
	return strings.Replace(strings.Replace(str, "\n", "", -1), "\t", "", -1)
}

type reqBodyLogReader struct {
	io.ReadCloser
	buffer *bodyBuffer
}

func (r *reqBodyLogReader) Read(p []byte) (n int, err error) {
	b, err := ioutil.ReadAll(r.ReadCloser)
	if nil != err {
		return 0, err
	}
	oldReader := r.ReadCloser
	defer func(oldReader io.ReadCloser) {
		_ = oldReader.Close()
	}(oldReader)

	go func() {
		_, _ = r.buffer.Write(b)
	}()

	r.ReadCloser = ioutil.NopCloser(bytes.NewReader(b))
	return r.ReadCloser.Read(p)
}

func (r *reqBodyLogReader) Close() error {
	return r.ReadCloser.Close()
}

type respBodyLogWriter struct {
	gin.ResponseWriter
	buffer *bodyBuffer
}

func (w *respBodyLogWriter) Write(b []byte) (int, error) {
	_, _ = w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}

const HTTPRequestStartTime = "httpRequestStartTime"

// 日志中间件
func Logger(hostPrefix string, logger *logger.Logger, notLogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}
	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	return func(ctxt *gin.Context) {
		start := date.Now()
		ctxt.Set(HTTPRequestStartTime, start)
		reqReader := &reqBodyLogReader{buffer: &bodyBuffer{Buffer: *bytes.NewBufferString("")}, ReadCloser: ctxt.Request.Body}
		ctxt.Request.Body = reqReader
		respWriter := &respBodyLogWriter{buffer: &bodyBuffer{Buffer: *bytes.NewBufferString("")}, ResponseWriter: ctxt.Writer}
		ctxt.Writer = respWriter

		ctxt.Next()

		if ctxt.Writer.Status() < http.StatusBadRequest { // httpStatus大于等于400的不应在此记录，而应该panic抛给下面的Recovery方法处理
			path := ctxt.Request.URL.Path
			if _, ok := skip[path]; !ok {
				end := date.Now()
				latency := end.Sub(start)

				httpStatus := ctxt.Writer.Status()
				clientIP := ctxt.ClientIP()
				req, fields := splitUri(hostPrefix, ctxt)
				comment := ctxt.Errors.ByType(gin.ErrorTypePrivate).String()
				logHttpRequest := true
				if i, ok := ctxt.Get(IsNoLogHTTPRequest); ok {
					if isNoLogHttpRequest, ok := i.(bool); ok && isNoLogHttpRequest {
						logHttpRequest = false
					}
				}
				if logHttpRequest {
					logger = logger.WithField("requestBody", reqReader.buffer.String())
				}
				if i, ok := ctxt.Get(IsLogHTTPResponse); ok {
					if isLogHttpResponse, ok := i.(bool); ok && isLogHttpResponse {
						logger = logger.WithField("responseBody", respWriter.buffer.String())
					}
				}
				logger.
					WithCaller(7).
					WithField("tag", "API").
					WithField("lib", "gin").
					WithField("httpStatus", httpStatus).
					WithField("latency", fmt.Sprintf("%v", latency)).
					WithField("clientIP", clientIP).
					WithField("comment", comment).
					WithFields(fields).
					Info(req)

			}
		}
	}
}
