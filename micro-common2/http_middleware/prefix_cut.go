package http_middleware

import (
	"net/http"
	"strings"
)

// 请求前缀剪裁中间件（底层http.Handler接口织入，非Gin中间件）
type PrefixCut struct {
	Handler    http.Handler
	HostPrefix string
}

func (cut *PrefixCut) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if len(strings.TrimSpace(cut.HostPrefix)) > 0 {
		hostPrefixPattern := getHostPrefixPattern(cut.HostPrefix)
		if hostPrefixPattern.MatchString(request.URL.Path) {
			request.URL.Path = hostPrefixPattern.ReplaceAllString(request.URL.Path, "")
			request.RequestURI = hostPrefixPattern.ReplaceAllString(request.RequestURI, "")
		}
	}
	cut.Handler.ServeHTTP(writer, request)
}
