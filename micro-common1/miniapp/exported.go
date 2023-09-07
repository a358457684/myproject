package miniapp

import (
	"common/config"
	"common/log"
	"errors"
)

func Init() {
	var err error
	defer func() {
		if err != nil {
			log.WithError(err).Error("微信小程序初始化失败")
			panic(err)
		}
		log.Info("微信小程序初始化成功")
	}()
	if config.Data.Miniapp == nil {
		err = errors.New("读取微信小程序配置失败")
		return
	}
	err = initMiniapp(config.Data.Miniapp)
}
