package utils

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"pp/common-golang/ginx"

	"github.com/gin-gonic/gin"
)

//easyjson:json
type Res struct {
	ResHead
	Body interface{} `json:"body,omitempty"`
}
type BanBanRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
	Status  bool        `json:"status"`
}

//easyjson:json
type ResHead struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg,omitempty"`
	Detail     string `json:"-"`
	HttpStatus int    `json:"-"`
}

func E0(errCode IErrorCode, args ...interface{}) *Res {
	errMsg := ReadErrMsg(errCode, args...)
	return E2(errCode.ToCode(), errCode.ToStatus(), errMsg, nil)
}
func ReadErrMsg(errCode IErrorCode, args ...interface{}) (errMsg string) {
	if len(args) < 1 {
		errMsg = errCode.GetDesc()
	} else {
		errMsg = errCode.ErrMsg(args...)
	}

	return
}

/**
  errDetail,记录到日志里面的详细信息
  errCode,规范日志记录，包含code,status,formatMessage
  args，formatMessage格式化日志输出参数
*/
func E(errDetail interface{}, errCode IErrorCode, args ...interface{}) *Res {
	errMsg := ReadErrMsg(errCode, args...)
	return E2(errCode.ToCode(), errCode.ToStatus(), errMsg, errDetail)
}

func E2(code, status int, errMsg string, errDetail interface{}) *Res {
	var msg, detail string
	msg = strings.TrimSpace(errMsg)
	if nil != errDetail {
		if err, ok := errDetail.(error); ok {
			detail = fmt.Sprintf("%s %s", msg, err.Error())
		} else {
			detail = strings.TrimSpace(fmt.Sprintf("%s", errDetail))
		}
	}
	result := &Res{
		ResHead: ResHead{
			Code:   code,
			Msg:    msg,
			Detail: detail,
		},
		Body: errDetail,
	}

	if status > 0 {
		result.HttpStatus = status
	} else {
		result.HttpStatus = http.StatusBadRequest
	}
	return result
}
func BanBanE(code int, errMsg string) *BanBanRes {
	result := &BanBanRes{
		Code:    code,
		Message: strings.TrimSpace(errMsg),
		Result:  nil,
		Status:  false,
	}
	return result
}
func BanBanR(body interface{}) *BanBanRes {
	return &BanBanRes{
		Code:    http.StatusOK,
		Message: "success",
		Result:  body,
		Status:  true,
	}
}

func R1(ctx *gin.Context, body interface{}) *Res {
	v, err := jsoniter.Marshal(body)
	if err == nil {
		ginx.SetResBody(ctx, v)
	}

	return R(body)

}

func R(body interface{}) *Res {
	return &Res{
		ResHead: ResHead{
			Code:       http.StatusOK,
			Msg:        "success",
			HttpStatus: http.StatusOK,
		},
		Body: body,
	}
}

func (v *Res) Write(c *gin.Context) {

	if v.HttpStatus >= 400 || v.HttpStatus < 600 {
		panic(c)
		return
	}

	c.JSON(v.HttpStatus, v)
	c.Abort()
}

func (v *Res) Error() string {
	return v.Msg
}
func (v *BanBanRes) Error() string {
	return v.Message
}

// 获取客户端IP地址
func GetRemoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("Remote_addr"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}
	return remoteAddr
}
