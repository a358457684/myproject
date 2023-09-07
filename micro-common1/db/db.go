package db

import (
	"common/config"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//读
var RDB *sqlx.DB

//写
var WDB *sqlx.DB

//初始化数据库客户端， 实现读写分离，对部分参数进行设置
func initDB(dboptions *config.DBOptions) error {

	//初始化读
	var err error
	WDB, err = initDBEntry(dboptions.Master)
	if err != nil {
		return err
	}

	if dboptions.Slave != nil {
		RDB, err = initDBEntry(dboptions.Slave)
	} else {
		RDB = WDB
	}
	return nil
}

func initDBEntry(dbentry *config.DBOptionsEntry) (*sqlx.DB, error) {
	if dbentry == nil {
		return nil, errors.New("数据库配置为空")
	}
	info := fmt.Sprintf("%s:%s@%s/%s?%s", dbentry.Username, dbentry.Password, dbentry.Host, dbentry.Path, dbentry.RawQuery)
	db, err := sqlx.Connect(dbentry.Dialector, info)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(dbentry.MaxOpen)
	db.SetMaxIdleConns(dbentry.MaxIdle)
	return db, nil
}
