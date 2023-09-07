package restful

import (
	"common/biz/dto"
	"common/util"
)

const (
	ApplyRobotJobUrl     = "/api/applyJob"          //接收任务
	RobotJobCompletedUrl = "/api/completeJob"       //任务完成
	RobotJobArrivedUrl   = "/api/jobArrived"        //任务到达
	ApplyDisinfectJobUrl = "/api/applyDisinfectJob" //接收消毒任务路径
	PingUrl              = "/ping"                  //ping接口
)

//ApplyRobotJob 申请任务.
func ApplyRobotJob(serverAddr string, data dto.ApplyRobotJob) error {
	if err := Post(serverAddr+ApplyRobotJobUrl, data, nil); err != nil {
		return util.WrapErr(err, "任务申请失败")
	}
	return nil
}

//ApplyDisinfectJob 申请消毒任务.
func ApplyDisinfectJob(serverAddr string, data dto.ApplyDisinfectJob) error {
	if err := Post(serverAddr+ApplyDisinfectJobUrl, data, nil); err != nil {
		return util.WrapErr(err, "任务申请失败")
	}
	return nil
}

//RobotJobArrived 任务到达.
func RobotJobArrived(serverAddr string, data dto.RobotJobArrived) error {
	if err := Put(serverAddr+RobotJobArrivedUrl, data, nil); err != nil {
		return util.WrapErr(err, "任务到达失败")
	}
	return nil
}

//RobotJobCompleted 任务完成.
func RobotJobCompleted(serverAddr string, data dto.RobotJobCompleted) error {
	if err := Delete(serverAddr+RobotJobCompletedUrl, data, nil); err != nil {
		return util.WrapErr(err, "完成任务失败")
	}
	return nil
}
