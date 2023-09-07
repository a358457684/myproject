package monitor_mqtt

import (
	"micro-common1/biz/enum"
	"time"
)

// topic
const (

	// 服务端
	toServer = "/airobot/to_service/+/+/+"

	// 客户端
	toClient = "/airobot/to_robot/%s/%s/%s"

	// 代理服务
	toProxy = "/airobot/to_proxy/%s/%s/%s"

	// 发布机器人pad端(Y2R、E2R)
	padToOffice = "/pad/toOffice/"

	// 发布机器人pad端(Y2P)
	toPad = "/airobot/to_pad/%s/%s/%s"

	// 任务取消
	jobCancel = "jobCancle"

	// 收到返回状态: 0成功, 非0失败
	feedbackSucceed = "0"
)

// 服务端发送的消息结构 TODO
type serverMsgVo struct {
	Path     string      `json:"path,omitempty"`  // 老版
	Body     interface{} `json:"body,omitempty"`  // 老版
	MsgId    string      `json:"msgId,omitempty"` // 老版
	Cmd      int         `json:"cmd,omitempty"`
	DispID   int64       `json:"dispID,omitempty"`
	JobGroup interface{} `json:"jobGroup,omitempty"`
}

// MQTT 发送 回复
type MqttMsgVo struct {
	Cmd           string             `json:"cmd,omitempty"`    // 回复
	Status        int                `json:"status,omitempty"` // 回复
	Token         string             `json:"token,omitempty"`
	Origin        enum.MsgOriginEnum `json:"origin,omitempty"`
	MsgID         int64              `json:"msgID,omitempty"` // 发送 回复 required
	Data          interface{}        `json:"data,omitempty"`
	FirstSendTime string             `json:"firstSendTime,omitempty"`
	LastSendTime  string             `json:"lastSendTime,omitempty"`
	Remark        string             `json:"remark,omitempty"`
}

type MqttMessageData struct {
	Code     int         `json:"code"`
	MqttType string      `json:"type"`
	MsgId    string      `json:"msgId"`
	Time     int64       `json:"time"`
	Data     interface{} `json:"data"`
}

type Position struct {
	Orientation float64 `json:"orientation"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
}

// 机器人发过来的状态字段
type RobotStatusMessage struct {
	DisfectionRemainTime int                  `json:"disfectionRemainTime"` // pad端有用
	Electric             float64              `json:"electric"`
	EstopStatus          int                  `json:"estopStatus"`
	Floor                int                  `json:"floor"`
	Halt                 string               `json:"halt"`
	InitSpotName         string               `json:"initSpotName"`
	JobType              enum.JobTypeEnum     `json:"jobType"`
	LastStatus           int                  `json:"lastStatus"`
	PauseType            int                  `json:"pauseType"`
	PersonDetect         int                  `json:"personDetect"` // pad端有用
	Position             Position             `json:"position"`
	SoftEstopStatus      int                  `json:"softEstopStatus"` // pad端有用
	SpotId               string               `json:"spotId"`
	SpotName             string               `json:"spotName"`
	Status               enum.RobotStatusEnum `json:"status"`
	Timestamp            int                  `json:"timestamp"`
	WaterStatus          int                  `json:"waterStatus"` // pad端有用
	BaseObjId            int                  `json:"baseObjId"`   // pad端有用
	JobId                string               `json:"jobId"`
	Target               string               `json:"target"`
	TargetName           string               `json:"targetName"`
	GroupId              string               `json:"groupId"`
}

// 回复结构体
type feedback struct {
	Cmd    string `json:"cmd"`
	MsgID  int64  `json:"msgID"`
	Status int    `json:"status"`
}

type commandStatus struct {
	MsgId       string      `json:"msgId"`
	OfficeId    string      `json:"officeId"`
	UserId      string      `json:"userId"`      // 用户ID，可以是frontServerId,或者是robotId
	Path        string      `json:"path"`        // 消息path 例如 ippbx
	Msg         interface{} `json:"msg"`         // 消息内容
	SendCount   int         `json:"sendCount"`   // 发送次数，默认为1次
	SendTime    int64       `json:"sendTime"`    // 最后一次发送时间戳
	CreateTime  int64       `json:"createTime"`  // 消息最早创建时间
	Status      int         `json:"status"`      // 状态；0-已发送，未返回 1-失败
	UseDispatch bool        `json:"useDispatch"` // 是否使用了调度系统的
	NetType     int         `json:"netType"`     // 网络连接类型
}

type terminalJobVo struct {
	TerminalNo string `json:"terminalNo"`
	Status     int    `json:"status"`  // 1:队列中，2:执行中，3:已到达，4:已完成，5:失败
	JobId      string `json:"jobId"`   // 任务id
	RobotId    string `json:"robotId"` // 机器人id
	CountDown  int    `json:"countDown"`
	Rank       int    `json:"rank"` // 排名
}

type ProxyServer struct {
	OfficeId   string    `json:"officeId,omitempty"`
	ProxyId    string    `json:"proxyId,omitempty"`
	UploadTime time.Time `json:"uploadTime,omitempty"`
}
