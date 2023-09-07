package orm

import (
	"common/config"
	"common/log"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
	"net/url"
	"strings"
	"time"
)

var DB *gorm.DB

func initDB(options *config.DBOptions) error {
	log.Info("开始初始化数据库...")

	masterDialector, err := getDialector(options.Master)
	if err != nil {
		return err
	}
	DB, err = gorm.Open(masterDialector, &gorm.Config{
		Logger:      &Logger{LogLevel: logger.Info},
		PrepareStmt: true,
		// Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	if err != nil {
		return err
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetConnMaxIdleTime(time.Hour)
	sqlDB.SetConnMaxLifetime(time.Hour * 24)
	sqlDB.SetMaxOpenConns(max(options.Master.MaxOpen, 10))
	sqlDB.SetMaxIdleConns(max(options.Master.MaxIdle, 20))

	if options.Slave == nil {
		return nil
	}
	log.Info("开始初始化从数据库...")
	slaveDialector, err := getDialector(options.Slave)
	dbResolver := dbresolver.Register(dbresolver.Config{
		Replicas: []gorm.Dialector{slaveDialector},
	}).
		SetConnMaxIdleTime(time.Hour).
		SetConnMaxLifetime(time.Hour * 24).
		SetMaxOpenConns(max(options.Slave.MaxOpen, 10)).
		SetMaxIdleConns(max(options.Slave.MaxIdle, 20))
	if err := DB.Use(dbResolver); err != nil {
		return err
	}
	return nil
}

func getDialector(options *config.DBOptionsEntry) (gorm.Dialector, error) {
	u := url.URL{
		Scheme:   options.Dialector,
		User:     url.UserPassword(options.Username, options.Password),
		Host:     options.Host,
		Path:     options.Path,
		RawQuery: options.RawQuery,
	}
	var dialector gorm.Dialector
	switch options.Dialector {
	case "sqlite":
		dialector = sqlite.Open(u.Path)
	case "mysql":
		if !strings.Contains(u.Host, "tcp") {
			u.Host = fmt.Sprintf("tcp(%s)", u.Host)
		}
		dialector = mysql.Open(u.String()[8:])
	case "sqlserver":
		dialector = sqlserver.Open(u.String())
	default:
		return nil, errors.New("未知的数据库配置模式")
	}
	log.Info(u.String())
	log.Infof("数据库模式：%s", options.Dialector)
	return dialector, nil
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
