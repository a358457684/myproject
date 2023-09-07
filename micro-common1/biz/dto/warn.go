package dto

import "common/biz/enum"

type Warn struct {
	OfficeID  string             `json:"officeID"`          // 机构ID
	RobotID   string             `json:"robotID"`           // 机器人ID
	GUID      string             `json:"guId,omitempty"`    // 位置GUID
	Origin    enum.MsgOriginEnum `json:"origin"`            // 来源
	Message   string             `json:"Message,omitempty"` // 发送消息内容
	ErrorCode int                `json:"errorCode"`         // 错误码
	SendTime  int64              `json:"sendTime"`          // 发送时间
}
