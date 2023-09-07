package dto

type Mail struct {
	MailTo    []string `json:"mailTo"`    // 收件人
	Subject   string   `json:"subject"`   // 标题
	Body      string   `json:"body"`      // 内容
	AnnexPath string   `json:"annexPath"` // 附件路径
	AnnexName string   `json:"annexName"` // 附件名
}

type Phone struct {
	TelNumbers []string `json:"telNumbers"` // 电话号码, 传入号码切片， 例如 []string{"18888888888","13232323232"}
	VoiceText  string   `json:"voiceText"`  // 文本内容
	TtsCode    string   `json:"ttsCode"`    // 语音模板号，由调用方提供
}

type WeChat struct {
	ToUser     []string `json:"touser"`        // 必须, 接受者OpenID
	TemplateId string   `json:"template_id"`   // 必须, 模版ID
	URL        string   `json:"url,omitempty"` // 可选, 用户点击后跳转的URL, 该URL必须处于开发者在公众平台网站中设置的域中

	AppId    string `json:"appid"`    // 可选; 所需跳转到的小程序appid（该小程序appid必须与发模板消息的公众号是绑定关联关系）
	PagePath string `json:"pagepath"` // 可选; 跳转小程序的页面路径

	Data interface{} `json:"data"` // 必须, 模板数据, struct 或者 *struct, encoding/json.Marshal 后满足格式要求.
}

// OK	请求成功
// isp.RAM_PERMISSION_DENY	RAM权限DENY
// isv.OUT_OF_SERVICE	业务停机
// isv.PRODUCT_UN_SUBSCRIPT	未开通云通信产品的阿里云客户
// isv.PRODUCT_UNSUBSCRIBE	产品未开通
// isv.ACCOUNT_NOT_EXISTS	账户不存在
// isv.ACCOUNT_ABNORMAL	账户异常
// isv.VOICE_FILE_ILLEGAL	语音文件不合法
// isv.BILLID_NOT_EXIST	号显不合法
// isv.INVALID_PARAMETERS	参数异常
// isp.SYSTEM_ERROR	系统错误
// isv.MOBILE_NUMBER_ILLEGAL	号码格式非法
// isv.BUSINESS_LIMIT_CONTROL	触发流控
// CALL_ERROR  本地调用错误
// NULL_RESPONSE 空回复
// NULL_CODE 空编码
type CallResponse struct {
	Code    string `json:"code"`    // 语音返回错误码 ，参考上面错误描述
	Message string `json:"message"` // 语音返回错误信息
}

// 微信返回信息
type WechatResponse struct {
	Code    string   `json:"code"`    // http 错误码
	Message string   `json:"message"` // http错误提示
	Data    []string `json:"data"`    // 错误信息内容，顺序与发送的用户一致，空字符串表示成功
}
