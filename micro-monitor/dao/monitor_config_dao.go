package dao

import (
	"epshealth-airobot-monitor/model"
	"micro-common1/orm"
)

// monitorType 监控类型（1：任务，2：范围）默认为1
func GetByOfficeIdAndRobotIdAndMonitorType(officeId string, robotId string, monitorType string) model.JobScopeMonitorConfig {
	sql := `SELECT DISTINCT
				a.office_id,
				a.monitor_scope,
				a.monitor_type,
				a.robot_status,
				a.robot_id
			FROM device_job_scope_monitor_config a
			WHERE a.office_id = ?
				and a.robot_id = ?
				and a.monitor_type = '1'
				and a.del_flag = '0'
			LIMIT 1`
	var jobConfig model.JobScopeMonitorConfig
	orm.DB.Raw(sql, officeId, robotId, monitorType).Scan(&jobConfig)
	return jobConfig
}
