package dto

type Result struct {
	Code    int         `json:"code"`    //错误编码
	Data    interface{} `json:"data"`    //数据
	Message string      `json:"message"` //描述信息
	Error   string      `json:"error"`   //错误信息
}
