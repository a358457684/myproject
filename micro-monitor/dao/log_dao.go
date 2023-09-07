package dao

import "micro-common1/orm"

const (
	// 日志类型（1：正常日志；2：错误日志）
	TypeAccess    = "1"
	TypeException = "2"

	// 系统类型（0：管理后台；1：监控系统）
	MonitorSystem = 1
)

type Log struct {
	Id         string
	LogType    string `gorm:"column:type"` // 日志类型
	Title      string // 日志标题
	RemoteAddr string // 地址
	UserAgent  string // 客户端
	RequestUri string // URI
	Method     string // 请求方法
	Params     string // 参数
	Exception  string // 异常信息
	SystemType int    // 日志记录系统类型
	CreateBy   string // 操作人
	CreateDate string // 操作时间
}

func (Log) TableName() string {
	return "sys_log"
}

func InsertLog(log Log) int64 {
	result := orm.DB.Create(&log)
	return result.RowsAffected
}
