package dao

import (
	"epshealth-airobot-monitor/utils"
	"micro-common1/biz/manager"
	"micro-common1/orm"
	"strings"
)

// 机器人
type Robot struct {
	Id                  string            `json:"id"`                  // 编号
	Name                string            `json:"name"`                // 名称
	Model               manager.RobotType `json:"model"`               // 型号
	Account             string            `json:"account"`             // 登录账号
	ChassisSerialNumber string            `json:"chassisSerialNumber"` // 序列号
	SoftVersion         string            `json:"softVersion"`         // 软件版本
	OfficeId            string            `json:"officeId"`            // 机构id
	OfficeName          string            `json:"officeName"`          // 机构id
}

func (Robot) TableName() string {
	return "device_robot"
}

func FindRobotByUser(officeId string, user utils.JwtData) (robots []Robot) {

	var sql strings.Builder
	sql.WriteString(`SELECT DISTINCT
							t.id,
							t.name,
							t.office_id,
							t2.name AS "OfficeName",
							t.account,
							t.model,
							t.chassis_serial_number,
							t.soft_version
						FROM
							device_robot AS t
							INNER JOIN sys_office AS t2 ON t.office_id = t2.id AND t2.del_flag = '0' `)

	// 不是全机构
	notAllOffice := !IsAdmin(user) && !user.HasAllOffice

	if notAllOffice {
		sql.WriteString(`INNER JOIN sys_user_office AS t3 ON t.office_id = t3.office_id `)
	}

	sql.WriteString(`WHERE t.del_flag = '0' `)

	var params []interface{}

	if notAllOffice {
		sql.WriteString(`AND t3.user_id = ? `)
		params = append(params, user.Id)
	}

	if officeId != "0" {
		sql.WriteString(`AND t.office_id = ? `)
		params = append(params, officeId)
	}

	orm.DB.Raw(sql.String(), params...).Scan(&robots)
	return
}

func GetByRobotId(id string) (robot Robot) {
	sql := `SELECT
				t1.id,
				t1.name,
				t1.model,
				t1.account,
				t1.chassis_serial_number,
				t1.soft_version,
				t2.id AS "OfficeId",
				t2.name AS "OfficeName"
			FROM
				device_robot AS t1
				INNER JOIN sys_office AS t2 ON t1.office_id = t2.id AND t2.del_flag = '0'
			WHERE 
				t1.del_flag = '0'
				AND t1.id = ?
			LIMIT 1`
	orm.DB.Raw(sql, id).Scan(&robot)
	return
}

func FindRobotModels() (models []string) {
	orm.DB.Table("device_robot_model").Select("model_value").Find(&models, "del_flag = '0'")
	return
}
