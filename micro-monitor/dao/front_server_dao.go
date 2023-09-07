package dao

import "micro-common1/orm"

type FrontServer struct {
	Id      string
	Ip      string
	Account string
}

func (FrontServer) TableName() string {
	return "device_front_server"
}

func FindFrontServerByOfficeId(officeId string) (dataList []FrontServer) {
	orm.DB.Find(&dataList, "del_flag = '0' and status = 1 and office_id = ?", officeId)
	return
}
