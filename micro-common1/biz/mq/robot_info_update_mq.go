package mq

import (
	"common/biz/dto"
	"common/log"
	"common/rabbitmq"
	"common/util"
	"encoding/json"
	"github.com/streadway/amqp"
	"time"
)

type RobotStatusRoute string

// 监控系统向医护端推送，机器人信息变化
const (
	// 机器人信息变化路由
	robotStatusRoute RobotStatusRoute = "routingkey_socket_robot_info.#"
)

var (
	// 机器人信息变化交换机
	RobotInfoExchange = rabbitmq.Exchange{
		Name:  "exchange_socket_robot_info",
		Model: rabbitmq.ET_Topic,
	}
	// 机器人信息变化队列
	RobotInfoQueue = "queue_socket_robot_info"
)

var (
	_robotInfoProduce *rabbitmq.RbProcedure
	_robotInfoConsume *rabbitmq.RbConsume
)

// 机器人信息变化es；机器人位置变化vo
type RobotStatusChangeMqVo struct {
	OldStatus dto.RobotStatus `json:"oldStatus"`
	NewStatus dto.RobotStatus `json:"newStatus"`
	SentTime  int64           `json:"sentTime"`
}

//机器人信息基本变更推送消息
func RobotStatusUpdate(robotStatusChangeVoinf RobotStatusChangeMqVo) error {
	bf := util.GetBuffer()
	err := json.NewEncoder(bf).Encode(robotStatusChangeVoinf)
	//buf, err := json.Marshal(jobinf)
	if err != nil {
		util.FreeBuffer(bf)
		log.WithError(err).Error("机器人信息基本变更序列化状态信息失败")
		return err
	}
	if _robotInfoProduce == nil {
		produce, err := rabbitmq.DefaultRMQ.RegisterProcedure(RobotInfoExchange)
		if err != nil {
			util.FreeBuffer(bf)
			log.WithError(err).Error("机器人信息基本变更生产者创建失败")
			return err
		}
		_robotInfoProduce = produce
	}
	err = _robotInfoProduce.PublishSimple(string(robotStatusRoute), bf.Bytes())
	util.FreeBuffer(bf)
	return err
}

//SubscribeRobotStatusUpdate 订阅机器人基本信息变更
func SubscribeRobotStatusUpdate(handler func(time time.Time, data RobotStatusChangeMqVo)) error {
	if _robotInfoConsume == nil {
		consume, err := rabbitmq.DefaultRMQ.RegisterConsume(RobotInfoExchange, "", true, func(d amqp.Delivery) {
			var changeVo RobotStatusChangeMqVo
			err := json.Unmarshal(d.Body, &changeVo)
			if err != nil {
				log.WithError(err).Error("接收机器人状态变更，数据解析失败")
				return
			}
			handler(time.Now(), changeVo)
		})
		if err != nil {
			return err
		}
		_robotInfoConsume = consume
		consume.Subscribe(string(robotStatusRoute))
	}
	return nil
}
