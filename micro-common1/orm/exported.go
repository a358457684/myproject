package orm

import (
	"common/config"
	"common/log"
)

func init() {
	if config.Data.DB == nil {
		log.Error("读取数据库配置失败")
		return
	}
	if config.Data.DB.UnUseOrm { //不启用orm
		return
	}
	err := initDB(config.Data.DB)
	if err != nil {
		log.WithError(err).Error("数据库初始化失败")
		panic(err)
	} else {
		log.Info("数据库初始化成功")
	}
}
