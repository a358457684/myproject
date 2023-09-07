package mq

import (
	"common/log"
	"common/rabbitmq"
	"common/util"
	"encoding/json"
	"github.com/streadway/amqp"
	"time"
)

// 监控系统向医护端推送，机器人位置变化
const (
	// 机器人信息变化路由
	robotPositionRoute RobotStatusRoute = "routingkey_socket_robot_position.#"
)

var (
	// 机器人位置变化交换机
	RobotPositionExchange = rabbitmq.Exchange{
		Name:  "exchange_socket_robot_position",
		Model: rabbitmq.ET_Topic,
	}
)

var (
	_robotPositionProduce *rabbitmq.RbProcedure
	_robotPositionConsume *rabbitmq.RbConsume
)

//机器人位置变更推送消息
func RobotPositionUpdate(robotStatusChangeVoinf RobotStatusChangeMqVo) error {
	bf := util.GetBuffer()
	err := json.NewEncoder(bf).Encode(robotStatusChangeVoinf)
	//buf, err := json.Marshal(jobinf)
	if err != nil {
		util.FreeBuffer(bf)
		log.WithError(err).Error("机器人位置变更序列化状态信息失败")
		return err
	}
	if _robotPositionProduce == nil {
		produce, err := rabbitmq.DefaultRMQ.RegisterProcedure(RobotPositionExchange)
		if err != nil {
			util.FreeBuffer(bf)
			log.WithError(err).Error("机器人位置变更生产者创建失败")
			return err
		}
		_robotPositionProduce = produce
	}
	err = _robotPositionProduce.PublishSimple(string(robotPositionRoute), bf.Bytes())
	util.FreeBuffer(bf)
	return err
}

//SubscribeRobotPositionUpdate 订阅机器人位置变更
func SubscribeRobotPositionUpdate(handler func(time time.Time, data RobotStatusChangeMqVo)) error {
	if _robotPositionConsume == nil {
		consume, err := rabbitmq.DefaultRMQ.RegisterConsume(RobotPositionExchange, "", true, func(d amqp.Delivery) {
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
		_robotPositionConsume = consume
		consume.Subscribe(string(robotPositionRoute))
	}
	return nil
}
