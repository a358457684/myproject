package dao

import (
	"epshealth-airobot-monitor/constant"
	"micro-common1/orm"
)

// 菜单
type MenuVo struct {
	Id         string `json:"id"`                      // 编号
	Name       string `json:"name"`                    // 名称
	MenuType   string `json:"type" gorm:"column:type"` // 是否在菜单中显示
	Permission string `json:"permission"`              // 权限标识
}

func (MenuVo) TableName() string {
	return "sys_menu"
}

func FindAllPermissions() []MenuVo {
	var entries []MenuVo
	orm.DB.Find(&entries, "del_flag = '0' AND permission LIKE CONCAT( ? , '%')", constant.MonitorPermission)
	return entries
}

func FindPermissionsByUserId(userId string) (entries []MenuVo) {
	sql := `SELECT DISTINCT
				a.id,
				a.name,
				a.type,
				a.permission
			FROM
				sys_menu a
				LEFT JOIN sys_menu p ON p.id = a.parent_id
				INNER JOIN sys_role_menu rm ON rm.menu_id = a.id
				INNER JOIN sys_role r ON r.id = rm.role_id AND r.useable = '1'
				INNER JOIN sys_user_role ur ON ur.role_id = r.id
				INNER JOIN sys_user u ON u.id = ur.user_id AND u.id = ?
			WHERE
				a.del_flag = '0'
				AND r.del_flag = '0'
				AND u.del_flag = '0'
				AND a.permission LIKE CONCAT( ? , '%' ) 
			ORDER BY
				a.sort`
	orm.DB.Raw(sql, userId, constant.MonitorPermission).Scan(&entries)
	return
}

func FindMenuByUser(permission string, userId string) (menuVo MenuVo) {
	sql := `SELECT DISTINCT
				a.id,
				a.name,
				a.type,
				a.permission
			FROM sys_menu AS a
			INNER JOIN sys_role_menu AS t2 ON a.id = t2.menu_id
			INNER JOIN sys_user_role AS t3 ON t2.role_id = t3.role_id
			WHERE
				a.del_flag = '0'
				AND a.permission = ?
				AND t3.user_id = ?
			LIMIT 1`
	orm.DB.Raw(sql, permission, userId).Scan(&menuVo)
	return
}
