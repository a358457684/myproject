package dao

import "micro-common1/orm"

func FindLiftIdsByOfficeId(officeId string) (ids []string) {
	orm.DB.Table("device_lift").Select("id").Find(&ids, "del_flag = '0' AND office_id = ?", officeId)
	return
}

// deviceSn 梯控模块序列号
func FindLiftDeviceSnById(id string) (deviceSn string) {
	orm.DB.Table("device_lift").Select("devicesn").Where(`del_flag = '0'`).Take(&deviceSn, id)
	return
}
