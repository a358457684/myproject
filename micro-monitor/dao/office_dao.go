package dao

import (
	"epshealth-airobot-monitor/model"
	"epshealth-airobot-monitor/utils"
	"micro-common1/biz/enum"
	"micro-common1/orm"
	"strings"
)

// 省市
type AreaVo struct {
	Id       string `json:"id"`       // 子id
	ParentId string `json:"parentId"` // 父id
	Name     string `json:"name"`     // 区域名称
}

// 机构
type OfficeVo struct {
	Id         string `json:"id"`         // 编号
	Name       string `json:"name"`       // 名称
	ProvinceId string `json:"provinceId"` // 省份
	CityId     string `json:"cityId"`     // 城市
}

type Office struct {
	OfficeVo
	model.BaseModel
}

type OfficeConfig struct {
	Id                 string
	OfficeId           string // 机构配置表父类
	RobotId            string // 机器人
	ErrorReturnTime    string // 错误返程时间（秒）
	ArrivalPendingTime string // 到达待接收时间（秒）
	IdleRechargeTime   string // 闲时充电时间（秒）
	ReceivingItemTime  string // 接收物品时间（秒）
	SelectBgmPath      string
	Remark             string                // 备注
	Mode               enum.DispatchModeEnum // 运行模式
	Logo               string
	SafeDistance       float64 // 安全距离
	DangerousDistance  float64 // 危险距离
	WorkTime           string  // 上班时间
	ClosedTime         string  // 下班时间
	GuId               string  // 上班位置
	ForceRechargeTime  string  // 强制充电时间
}

func (AreaVo) TableName() string {
	return "sys_area"
}

func (OfficeConfig) TableName() string {
	return "sys_office_config"
}

func (OfficeVo) TableName() string {
	return "sys_office"
}

//  1、2对应字典的省市类型
func FindAreaList() (areas []AreaVo) {
	orm.DB.Select("id", "parent_id", "name").Find(&areas, `del_flag = '0' and type in ("1", "2")`)
	return
}

func FindOffices(user utils.JwtData) (OfficeVos []OfficeVo) {
	var sql strings.Builder
	sql.WriteString(`SELECT
							t1.id,
							t1.name, 
							t2.id AS "ProvinceId",
							t3.id AS "CityId" 
						FROM
							sys_office AS t1 
							LEFT JOIN sys_area AS t2 ON t2.id = t1.province_area_id
							LEFT JOIN sys_area AS t3 ON t3.id = t1.city_area_id `)

	// 不是全机构
	notAllOffice := !IsAdmin(user) && !user.HasAllOffice

	if notAllOffice {
		sql.WriteString(`INNER JOIN sys_user_office AS t4 ON t1.id = t4.office_id `)
	}

	sql.WriteString(`WHERE t1.del_flag = '0' `)

	if notAllOffice {
		sql.WriteString(`AND t4.user_id = ? `)
	}

	sql.WriteString(`ORDER BY CONVERT ( t1.name USING gbk ) COLLATE gbk_chinese_ci ASC`)

	orm.DB.Raw(sql.String(), user.Id).Scan(&OfficeVos)
	return
}

func FindOfficeConfigByOffice(officeId string) (configs []OfficeConfig) {
	orm.DB.Select("id", "office_id", "robot_id", "mode").
		Find(&configs, "del_flag = '0' and office_id = ?", officeId)
	return
}

func FindAllOfficeConfigMode() (configs []OfficeConfig) {
	orm.DB.Select("id", "office_id", "robot_id", "mode").
		Find(&configs, "del_flag = '0' and mode != 0")
	return
}
