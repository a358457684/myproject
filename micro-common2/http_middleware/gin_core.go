package http_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

var tsPostfixPattern = regexp.MustCompile(`[?&](t=\d+&?)`) // 处理防重放防缓存的时间戳参数的正则

// 日志和错误处理需要记录日志中的请求信息处理
func splitUri(hostPrefix string, ctxt *gin.Context) (string, map[string]interface{}) {
	path := ctxt.Request.URL.Path
	method := ctxt.Request.Method
	raw := ctxt.Request.URL.RawQuery

	fields := make(map[string]interface{})
	uri := path
	if len(strings.TrimSpace(hostPrefix)) > 0 && !getHostPrefixPattern(hostPrefix).MatchString(path) {
		uri = strings.TrimSpace(hostPrefix) + "/" + strings.TrimLeft(path, "/")
	}
	query := raw
	if sub := tsPostfixPattern.FindStringSubmatch(query); nil != sub && len(sub) > 1 {
		query = strings.Replace(query, sub[1], "", -1)
	}
	if raw != "" {
		path = uri + "?" + raw
	}
	//if token := jws.GetToken(ctxt); len(token) > 0 {
	//	fields["token"] = token
	//}
	fields["endpoint"] = fmt.Sprintf("%s %s", method, uri) // GET /service/test/test
	fields["uri"] = uri                                    // /service/test/test
	fields["query"] = query                                // name=test&sex=1
	fields["method"] = method                              // GET
	req := fmt.Sprintf("%s %s", method, path)              // GET /service/test/test?name=test&sex=1&t=32523535323
	return req, fields
}

func getHostPrefixPattern(hostPrefix string) *regexp.Regexp {
	return regexp.MustCompile("^" + strings.TrimSpace(hostPrefix)) // 处理子系统路径前缀的正则
}
