package dao

import (
	"epshealth-airobot-monitor/constant"
	"epshealth-airobot-monitor/model"
	"micro-common1/biz/enum"
	"micro-common1/biz/manager"
	"micro-common1/orm"
	"strings"
	"time"
)

type RobotJobQueue struct {
	Id            string            `json:"id"`                         // 编号
	OfficeId      string            `json:"officeId"`                   // 机构
	RobotId       string            `json:"robotId"`                    // 机器人
	StartPosition string            `json:"startPosition"`              // 任务起始位置
	EndPosition   string            `json:"endPosition"`                // 任务结束位置
	StartUserId   string            `json:"startUserId"`                // 开启任务的用户
	JobType       enum.JobTypeEnum  `json:"jobType" gorm:"column:type"` // 任务类型
	Back          int               `json:"back"`                       // 是否立即返程
	GroupId       string            `json:"groupId"`                    // 任务分组
	JobOrder      int               `json:"jobOrder"`                   // 任务排序
	RobotModel    manager.RobotType `json:"robotModel"`                 // 机器人类型
	CreateDate    time.Time         `json:"createDate"`                 // 创建时间
	Remarks       string            `json:"remarks"`                    // 任务来源
}

func GetRobotJobById(id string) (robotJob model.RobotJob) {
	orm.DB.First(&robotJob, "del_flag = '0' and id = ?", id)
	return
}

func FindRobotJobByOffice(vo model.RobotJobQueryVo) (entries []model.RobotJob, total int64) {

	var sql strings.Builder
	sql.WriteString(`del_flag = '0' and office_id = ? and robot_id = ? `)

	params := []interface{}{vo.OfficeId, vo.RobotId}

	if vo.JobType != 0 {
		sql.WriteString(`and type = ? `)
		params = append(params, vo.JobType.Code())
	}

	if vo.JobStatus != 0 {
		sql.WriteString(`and status = ? `)
		params = append(params, vo.JobStatus.Code())
	}

	if vo.StartDate.IsZero() || vo.EndDate.IsZero() {
		sql.WriteString(`and status >= ? ORDER BY create_date DESC`)
		params = append(params, enum.JsCompleted.Code())
	} else {
		localStartData, _ := time.ParseInLocation(constant.DateTimeFormat, vo.StartDate.Format(constant.DateTimeFormat), time.Local)
		localEndData, _ := time.ParseInLocation(constant.DateTimeFormat, vo.EndDate.Format(constant.DateTimeFormat), time.Local)
		sql.WriteString(`and status >= ? and create_date >= ? and create_date <= ? ORDER BY create_date DESC`)
		params = append(params, enum.JsCompleted.Code(), localStartData, localEndData)
	}
	orm.DB.Limit(vo.PageSize).Offset((vo.PageIndex-1)*vo.PageSize).Where(sql.String(), params...).Find(&entries)
	orm.DB.Model(model.RobotJob{}).Where(sql.String(), params...).Count(&total)
	return
}

func SaveRobotJob(robotJob model.RobotJob) {
	orm.DB.Updates(&robotJob)
}

func SaveRobotJobLog(robotJobLog model.RobotJobLog) {
	orm.DB.Create(&robotJobLog)
}
