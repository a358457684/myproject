package model

type RobotWorkTimeConfigInfo struct {
	OfficeId         string // 机构ID
	RobotId          string // 机器人ID
	WorkConfigId     string // 工作配置ID
	IsPushSuccess    bool   // 是否推送成功
	IsPush           bool   // 是否推送
	PushTime         int64  // 推送时间
	PushSuccessTime  int64  // 推送成功时间
	PushPositionGuId string // 推送位置
	PushCount        int    // 推送次数
}
