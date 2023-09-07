package dao

import (
	"micro-common1/biz/enum"
	"micro-common1/orm"
	"time"
)

func UpdateDisinfectTaskLogEndTime(JobId string, status enum.JobStatusEnum) int64 {
	sql := `UPDATE device_disinfect_task_log 
				SET status = ? ,
				task_end_time = ? ,
				update_date = ? ,
				task_detail = ?  
			WHERE
				job_id = ?`
	res := orm.DB.Exec(sql, status.Code(), time.Now(), time.Now(), status.Message(), JobId)
	return res.RowsAffected
}
