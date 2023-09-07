package miniapp

import (
	"common/config"
	"errors"
	"github.com/chanxuehong/wechat/mp/oauth2"
)

var (
	AppId          = ""
	Secret         = ""
	Oauth2Endpoint *oauth2.Endpoint
)

func initMiniapp(options *config.MiniappOptions) error {
	if options == nil {
		return errors.New("微信小程序配置为空")
	}
	AppId = options.Appid
	Secret = options.Secret
	Oauth2Endpoint = oauth2.NewEndpoint(AppId, Secret)
	return nil
}
