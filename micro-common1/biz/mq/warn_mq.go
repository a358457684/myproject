package mq

import (
	"common/biz/dto"
	"common/log"
	"common/rabbitmq"
	"common/util"
	"encoding/json"
	"github.com/streadway/amqp"
)

const (
	warnRoute = "routing_warning"
	warnQueue = "queue_warning"
)

var (
	warnExchange = rabbitmq.Exchange{
		Name:  "exchange_warning",
		Model: rabbitmq.ET_Direct,
	}
	_warnProduce *rabbitmq.RbProcedure
	_warnConsume *rabbitmq.RbConsume
)

// 发布：预警消息
func WarnPub(data dto.Warn) (err error) {
	bf := util.GetBuffer()
	err = json.NewEncoder(bf).Encode(data)
	defer util.FreeBuffer(bf)
	if err != nil {
		log.WithError(err).Error("预警消息数据序列化失败")
		return
	}
	if _warnProduce == nil {
		if _warnProduce, err = rabbitmq.DefaultRMQ.RegisterProcedure(warnExchange); err != nil {
			log.WithError(err).Error("预警消息生产者创建失败")
			return
		}
	}
	err = _warnProduce.PublishSimple(warnRoute, bf.Bytes())
	if err != nil {
		log.WithError(err).Errorf("预警消息发送失败:%+v", data)
	}
	return
}

// 订阅：预警消息
func WarnSub(handler func(data dto.Warn)) (err error) {
	if _warnConsume == nil {
		_warnConsume, err = rabbitmq.DefaultRMQ.RegisterConsume(
			warnExchange,
			warnQueue,
			true,
			func(d amqp.Delivery) {
				var vo dto.Warn
				err := json.Unmarshal(d.Body, &vo)
				if err != nil {
					log.WithError(err).Error("接收预警，数据解析失败")
					return
				}
				handler(vo)
			})
		if err != nil {
			return
		}
		err = _warnConsume.Subscribe(warnRoute)
		if err != nil {
			log.WithError(err).Errorf("预警消息订阅失败")
		}
	}
	return
}
