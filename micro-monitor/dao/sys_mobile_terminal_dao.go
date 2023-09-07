package dao

import (
	"epshealth-airobot-monitor/model"
	"micro-common1/orm"
)

func FindMobileTerminalByOfficeId(officeId string) (entries []model.SysMobileTerminal) {
	orm.DB.Table("sys_mobile_terminal").Select("id").
		Find(&entries, "del_flag = '0' AND office_id = ?", officeId)
	return
}
