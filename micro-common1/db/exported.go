package db

import (
	"common/config"
	"common/log"
	"errors"
)

func Init() {
	var err error
	defer func() {
		if err != nil {
			log.WithError(err).Error("数据库初始化失败")
			panic(err)
		}
		log.Info("数据库初始化成功")
	}()
	if config.Data.DB == nil {
		err = errors.New("读取数据库配置失败")
		return
	}
	err = initDB(config.Data.DB)
}
