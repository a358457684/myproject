module common/orm

go 1.18

require (
	common/config v0.0.0
	common/log v0.0.0
	github.com/sirupsen/logrus v1.8.0
	gorm.io/driver/mysql v1.0.4
	gorm.io/driver/sqlite v1.1.4
	gorm.io/driver/sqlserver v1.0.6
	gorm.io/gorm v1.20.12
	gorm.io/plugin/dbresolver v1.1.0
)

replace common/config => ../config

replace common/log => ../log
