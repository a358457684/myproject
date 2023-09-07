package dao

import (
	"epshealth-airobot-monitor/utils"
	"micro-common1/orm"
)

// 系统用户
type User struct {
	Id           string // 编号
	LoginName    string // 登录名
	Password     string // 密码
	Name         string // 姓名
	HasAllOffice bool   // 是否关联所有机构
	LoginFlag    string // 是否可登录
}

// 机器人用户
type RobotUser struct {
	Id       string
	Username string // 用户名
	Nickname string // 呢称
}

func (User) TableName() string {
	return "sys_user"
}

func (RobotUser) TableName() string {
	return "device_robot_user"
}

func GetUser(username string) (user User) {
	orm.DB.First(&user, "del_flag = '0' and login_name = ? ", username)
	return
}

func IsAdmin(user utils.JwtData) bool {
	return user.Id == "1" && user.Username == "admin"
}

func FindRobotUserByOfficeId(officeId string) (robotUsers []RobotUser) {
	orm.DB.Find(&robotUsers, "del_flag = '0' and office_id = ?", officeId)
	return
}
