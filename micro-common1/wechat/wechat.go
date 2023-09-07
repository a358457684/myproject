package wechat

import (
	"common/config"
	"errors"
	"github.com/chanxuehong/wechat/mp/core"
)

var (
	AppId             = ""
	secret            = ""
	Token             = ""
	AesKey            = ""
	arriveTemplateId  = ""
	recieveTemplateId = ""
	miniAppId         = ""
	accessTokenServer core.AccessTokenServer
	WechatClient      *core.Client
)

func initwechat(options *config.WechatOptions) error {
	if options == nil {
		return errors.New("微信配置为空")
	}

	AppId = config.Data.Wechat.Appid
	secret = config.Data.Wechat.Secret
	Token = options.Token
	AesKey = options.AesKey
	arriveTemplateId = config.Data.Wechat.ArriveTemplateId
	recieveTemplateId = config.Data.Wechat.RecieveTemplateId
	miniAppId = config.Data.Wechat.MiniappId

	accessTokenServer = core.NewDefaultAccessTokenServer(AppId, secret, nil)
	WechatClient = core.NewClient(accessTokenServer, nil)

	return nil
}
