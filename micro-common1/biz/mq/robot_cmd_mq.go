package mq

import (
	"common/biz/enum"
	"common/mqtt"
	"common/util"
	"fmt"
	"time"
)

type RobotCmdBody struct {
	Origin        enum.MsgOriginEnum `json:"origin"`        //消息来源
	MsgID         int64              `json:"msgID"`         //消息ID
	SendTimes     int                `json:"sendTimes"`     //发送次数
	FirstSendTime util.JsonTime      `json:"firstSendTime"` //消息生成时间
	LastSendTime  util.JsonTime      `json:"lastSendTime"`  //最后发送时间
	Data          interface{}        `json:"data"`          //消息内容
	Remark        string             `json:"remark"`        //描述信息
}

type RobotCmdDTO struct {
	RobotCmdBody

	Cmd      enum.RobotCmdEnum `json:"cmd"`      //操作类型
	OfficeID string            `json:"officeID"` //机构ID
	RobotID  string            `json:"robotID"`  //机器人ID
}

func NewRobotCmdDTO(officeID, robotID string, cmd enum.RobotCmdEnum, data interface{}, origin enum.MsgOriginEnum, remark string) *RobotCmdDTO {
	return &RobotCmdDTO{
		RobotCmdBody: RobotCmdBody{
			MsgID:  util.GetSnowflakeID(),
			Data:   data,
			Remark: remark,
			Origin: origin,
		},
		Cmd:      cmd,
		OfficeID: officeID,
		RobotID:  robotID,
	}
}

//SendRobotCmd 给机器人下发操控指令.
func SendRobotCmd(operation *RobotCmdDTO) error {
	operation.SendTimes++
	now := time.Now()
	if operation.SendTimes == 1 {
		operation.FirstSendTime = util.JsonTime{Time: now}
	}
	operation.LastSendTime = util.JsonTime{Time: now}
	topic := fmt.Sprintf("/airobot/to_robot/%s/%s/%s", operation.OfficeID, operation.RobotID, operation.Cmd.Code())
	return mqtt.PublishCustom(topic, 2, false, operation.RobotCmdBody)
}
