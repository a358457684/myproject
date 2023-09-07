package dao

import (
	"micro-common1/orm"
)

type PositionConfig struct {
	PositionConfigReqVo
	// 下面参数暂时未使用到，如果使用的话，需要注意能否与数据库对应
	ArriveNotice           string // 到达提醒是否发送1:是，0:否
	ArriveNoticeUsers      string // 到达提醒接收的的机器人用户,用户id，逗号分隔
	ArriveNoticeStartUser  string // 到达提醒是否通知发送人1:是，0:否
	ArriveNoticeDetails    string // 到达提醒内容
	ArriveNoticeRemark     string // 到达提醒备注信息
	ReceiveNotice          string // 接收提醒是否发送1:是，0:否
	ReceiveNoticeUsers     string // 接收提醒接收的的机器人用户,用户id，逗号分隔
	ReceiveNoticeStartUser string // 接收提醒是否通知发送人1:是，0:否
	ReceiveNoticeDetails   string // 接收提醒内容
	ReceiveNoticeRemark    string // 接收提醒备注信息
}

type PositionConfigReqVo struct {
	OfficeId       string `json:"officeId" gorm:"column:office_id" binding:"required"`     // 机构ID
	RobotId        string `json:"robotId" gorm:"column:robot_id" binding:"required"`       // 机器人ID
	GuId           string `json:"guId" gorm:"column:gu_id" binding:"required"`             // 机器人中的位置guId
	NoticeType     string `json:"noticeType" gorm:"column:notice_type" binding:"required"` // 提醒类型 0-不通知、1-电话通知、2-声光通知、3-云端电话通知
	TeleNumber     string `json:"teleNumber" gorm:"column:tele_number"`                    // 分机号码
	SoundLightType string `json:"soundLightType" gorm:"column:sound_light_type"`           // 声光类型 1-(均匀声+闪灯) 2-(急促声+闪灯) 3-均匀声 4-急促声 5-闪灯
	Delay          string `json:"delay"`                                                   // 提醒时长
	SoundLightSn   string `json:"soundLightSn" gorm:"column:sound_light_sn"`               // 声光模块序列号
	Sensitivity    string `json:"sensitivity"`                                             // 灵敏度
	DtmfNumber     string `json:"dtmfNumber" gorm:"column:dtmf_number"`                    // 拨号号码
	VoiceText      string `json:"voiceText" gorm:"column:voice_text"`                      // 生成语音的配置文字
	SoundLightVer  string `json:"soundLightVer" gorm:"column:sound_light_ver"`             // 声光模块版本号
	Volume         int    `json:"volume"`                                                  // 音量
	AudioType      int    `json:"audioType" gorm:"column:audio_type"`                      // 语音提醒类型 (0-取消，1-语音+闪灯，2-语音+快速闪灯，3-语音，4-闪灯，5-快速闪灯，6-灯长亮）
	AudioFileNo    int    `json:"audioFileNo" gorm:"column:audio_file_no"`                 // 音频文件序列（1-15)
}

func (PositionConfig) TableName() string {
	return "device_robot_position_config"
}

// 根据机构Id、guId、提醒类型查询位置提醒
func (f *PositionConfig) FindWxNoticeList() []PositionConfig {
	positionConfigList := make([]PositionConfig, 0, 2)
	orm.DB.Find(&positionConfigList, "office_id=? and gu_id =? and notice_type=? and del_flag=0", f.OfficeId, f.GuId, f.NoticeType)
	return positionConfigList
}
