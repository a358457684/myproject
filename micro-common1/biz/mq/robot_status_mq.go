package mq

import (
	"common/biz/enum"
	"common/log"
	"common/rabbitmq"
	"common/util"
	"encoding/json"
	"github.com/streadway/amqp"
	"time"
)

type RobotStatusStatisticsRoute string

// 监控系统向医护端推送，机器人状态（任务记录）、机器人信息、与位置信息
const (
	// 任务执行记录路由
	robotStatusStatisticsRoute RobotStatusStatisticsRoute = "#.#.#"
)

var (
	// 任务执行记录交换机
	RobotStatusStatisticsExchange = rabbitmq.Exchange{
		Name:  "exchange_robot_status",
		Model: rabbitmq.ET_Topic,
	}
	// 任务执行记录队列
	RobotStatusStatisticsQueue = "queue_status_statistics"
)

var (
	_robotStatusProduce *rabbitmq.RbProcedure
	_robotStatusConsume *rabbitmq.RbConsume
)

// mq推送过来的机器人状态,旧的状态和新状态
type RobotStatusMqVo struct {
	RobotId   string            `json:"robotId"`
	OldStatus RobotStatusUpload `json:"oldStatus"`
	NewStatus RobotStatusUpload `json:"newStatus"`
	SentTime  int64             `json:"sentTime"`
}
type RobotStatusUpload struct {
	DocumentId      string                `json:"documentId"`      // 文档id
	RobotId         string                `json:"robotId"`         // 机器人id
	RobotAccount    string                `json:"robotAccount"`    // 机器人账号id
	OfficeId        string                `json:"officeId"`        // 机构id
	RobotModel      string                `json:"robotModel"`      // 机器人型号
	BuildingId      string                `json:"buildingId"`      // 楼宇id
	Status          enum.RobotStatusEnum  `json:"status"`          // 机器人状态
	JobId           string                `json:"jobId"`           // 任务id
	GroupId         string                `json:"groupId"`         // 任务组id
	LastUploadTime  int64                 `json:"lastUploadTime"`  // 最后上传时间
	X               float64               `json:"x"`               // x
	Y               float64               `json:"y"`               // y
	Z               float64               `json:"z"`               // z
	Orientation     float64               `json:"orientation"`     // 方位
	SpotId          string                `json:"spotId"`          // 最后位置
	Target          string                `json:"target"`          // 目标位置
	Process         []string              `json:"process"`         // 线路规则中要经过的位置
	NextSpot        string                `json:"nextSpot"`        // 下一个位置
	Floor           int                   `json:"floor"`           // 目标位置
	Electric        float64               `json:"electric"`        // 电量
	NetStatus       enum.NetStatusEnum    `json:"netStatus"`       // 网络状态
	JobType         int                   `json:"jobType"`         // 任务类型
	PauseType       int                   `json:"pauseType"`       // 是否暂停：1：暂停，0：正常（调度执行状态）
	EstopStatus     int                   `json:"estopStatus"`     // 是否急停（机器执行状态  0-正常，1-急停状态
	StatusStartTime int64                 `json:"statusStartTime"` // 状态开始时间
	StatusEndTime   int64                 `json:"statusEndTime"`   // 状态结束时间
	ExecState       enum.ExecStateEnum    `json:"execStateEnum"`   // 任务执行状态 // 注意看是否能解析成功
	TimeConsume     float64               `json:"timeConsume"`     // 耗时
	Message         string                `json:"message"`         // 网络状态
	FinalJobId      string                `json:"finalJobId"`      // 最终任务id
	DispatchMode    int                   `json:"dispatchMode"`    // 运行模式
	FinalJobType    int                   `json:"finalJobType"`    // 最终任务类型
	AcceptState     enum.AcceptStatusEnum `json:"acceptState"`     // 物品接收状态 0 待接收；1已接收；2超时未取
}

// 机器人任务状态变更推送消息
func RobotStatusStatistics(statusinf RobotStatusMqVo) error {
	bf := util.GetBuffer()
	err := json.NewEncoder(bf).Encode(statusinf)
	// buf, err := json.Marshal(jobinf)
	if err != nil {
		util.FreeBuffer(bf)
		log.WithError(err).Error("机器人任务状态变更序列化状态信息失败")
		return err
	}
	if _robotStatusProduce == nil {
		produce, err := rabbitmq.DefaultRMQ.RegisterProcedure(RobotStatusStatisticsExchange)
		if err != nil {
			util.FreeBuffer(bf)
			log.WithError(err).Error("机器人任务状态变更生产者创建失败")
			return err
		}
		_robotStatusProduce = produce
	}
	err = _robotStatusProduce.PublishSimple(string(robotStatusStatisticsRoute), bf.Bytes())
	util.FreeBuffer(bf)
	return err
}

// SubscribeRobotStatusStatistics 订阅机器人任务状态消息.
func SubscribeRobotStatusStatistics(handler func(time time.Time, data RobotStatusMqVo)) error {
	if _robotStatusConsume == nil {
		consume, err := rabbitmq.DefaultRMQ.RegisterConsume(RobotStatusStatisticsExchange, RobotStatusStatisticsQueue, true, func(d amqp.Delivery) {
			// jobStatuss := make([]RobotJobStatus, 0, 8)
			var robotStatusMqVo RobotStatusMqVo
			err := json.Unmarshal(d.Body, &robotStatusMqVo)
			if err != nil {
				log.WithError(err).Error("接收机器人任务状态变更，数据解析失败")
				return
			}
			handler(time.Now(), robotStatusMqVo)
		})
		if err != nil {
			return err
		}
		_robotStatusConsume = consume
		consume.Subscribe(string(robotStatusStatisticsRoute))
	}
	return nil
	/*return rabbitmq.Subscribe(RobotStatusStatisticsExchange, string(robotStatusStatisticsRoute), RobotStatusStatisticsQueue, true, func(message *rabbitmq.Message, delivery amqp.Delivery) {
		var robotStatusMqVo RobotStatusMqVo
		if err := message.UnMarshal(&robotStatusMqVo); err != nil {
			log.WithError(err).Error("接收机器人状态变更，数据解析失败")
			return
		}
		handler(message.Time, robotStatusMqVo)
	})*/
}
