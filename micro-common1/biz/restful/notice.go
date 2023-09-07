package restful

import (
	"common/biz/dto"
	"common/util"
)

const (
	MailUrl   = "/mail"   //发送邮件
	PhoneUrl  = "/phone"  //发送电话
	WechatUrl = "/wechat" //发送微信
)

//发送邮件
func SendMail(serverAddr string, vo dto.Mail) error {
	if err := Post(serverAddr+MailUrl, vo, nil); err != nil {
		return util.WrapErr(err, "邮件发送失败")
	}
	return nil
}

//发送电话
func SendPhone(serverAddr string, vo dto.Phone) (dto.Result, error) {
	var response dto.Result
	if err := Post(serverAddr+PhoneUrl, vo, &response); err != nil {
		return dto.Result{}, util.WrapErr(err, "电话发送失败")
	}
	return response, nil
}

//发送微信
func SendWechat(serverAddr string, vo dto.WeChat) error {
	var response dto.WechatResponse
	if err := Post(serverAddr+WechatUrl, vo, &response); err != nil {
		return util.WrapErr(err, "微信发送失败")
	}
	return nil
}
