package rabbitmq

import (
	"common/config"
	"common/log"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/suiyunonghen/DxCommonLib"
	"reflect"
)

var url string

var (
	DefaultRMQ *RabbitMQ
)

//编码器
type MQMsgCoder interface {
	Encode(data interface{}) ([]byte, error)
	Decode(msgdata []byte, destValue interface{}) error
}

func init() {
	if config.Data.RabbitMq == nil {
		log.Warn("跳过RabbitMQ初始化，读取RabbitMq配置失败")
		return
	}
	url = fmt.Sprintf("amqp://%s:%s@%s//%s", config.Data.RabbitMq.Username, config.Data.RabbitMq.Password, config.Data.RabbitMq.Service, config.Data.RabbitMq.Host)
	rmq, err := NewRabbitMQ(url, true, log.DefLogger())
	if err != nil {
		log.WithError(err).Error("rabbitmq启动失败")
		panic(err)
	}
	DefaultRMQ = rmq
}

func Subscribe(exchange Exchange, route, queue string, autoAck bool, handler func(*Message, amqp.Delivery)) error {
	consume, err := DefaultRMQ.RegisterConsume(exchange, queue, autoAck, func(delivery amqp.Delivery) {
		msg := &Message{}
		err := json.Unmarshal(delivery.Body, msg)
		if err != nil {
			log.WithError(err).Error("消息解析失败")
		}
		log.WithFields(logrus.Fields{
			"exchange":   delivery.Exchange,
			"routingKey": delivery.RoutingKey,
			"messageID":  msg.MessageID,
		}).Trace("接收到RabbitMq消息")
		handler(msg, delivery)
	})
	if err != nil {
		log.WithError(err).Error("消费者创建失败")
		return err
	}

	if err = consume.Subscribe(route); err != nil {
		log.WithError(err).Error("订阅消息失败")
		return err
	}
	return nil
}

var (
	StringType    = reflect.TypeOf((*string)(nil)).Elem()
	ByteArrayType = reflect.TypeOf((*[]byte)(nil)).Elem()
)

func SubscribeEx(exchange Exchange, route, queue string, autoAck bool, coder MQMsgCoder, destValue interface{}, handler func(interface{}, amqp.Delivery)) error {
	//consume, err := newConsume(url, exchange.Model, queue, exchange.Name, autoAck, true, log.DefLogger())
	destType := reflect.ValueOf(destValue).Type()
	if destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
	}
	consume, err := DefaultRMQ.RegisterConsume(exchange, queue, autoAck, func(delivery amqp.Delivery) {
		if coder != nil {
			destvalue := reflect.New(destType).Interface() //构建一个新的
			err := coder.Decode(delivery.Body, destvalue)
			if err != nil {
				log.WithError(err).Error("消息解析失败")
				return
			}
			handler(destvalue, delivery)
		} else {
			switch destType {
			case ByteArrayType:
				handler(nil, delivery)
			case StringType:
				handler(nil, delivery)
			default:
				destvalue := reflect.New(destType).Interface() //构建一个新的
				err := json.Unmarshal(delivery.Body, destValue)
				if err != nil {
					log.WithError(err).Error("消息解析失败，使用默认版本")
					handler(nil, delivery)
					return
				}
				handler(destvalue, delivery)
			}
		}
	})
	if err != nil {
		log.WithError(err).Error("消费者创建失败")
		return err
	}
	if err = consume.Subscribe(route); err != nil {
		log.WithError(err).Error("订阅消息失败")
		return err
	}
	return nil
}

func PublishEx(exchange Exchange, route string, data interface{}, coder MQMsgCoder) error {
	produre, err := DefaultRMQ.RegisterProcedure(exchange)
	if err != nil {
		log.WithError(err).Error("生产者创建失败")
		return err
	}
	if coder == nil {
		switch vdata := data.(type) {
		case []byte:
			produre.PublishSimple(route, vdata)
		case *[]byte:
			produre.PublishSimple(route, *vdata)
		case string:
			produre.PublishSimple(route, DxCommonLib.FastString2Byte(vdata))
		case *string:
			produre.PublishSimple(route, DxCommonLib.FastString2Byte(*vdata))
		default:
			bt, err := json.Marshal(data)
			if err != nil {
				log.WithError(err).Error("消息序列化编码失败")
				return err
			}
			produre.PublishSimple(route, bt)
		}
		return nil
	}
	bytes, err := coder.Encode(data)
	if err != nil {
		log.WithError(err).Error("消息序列化编码失败")
		return err
	}
	produre.PublishSimple(route, bytes)
	return nil
}

func Publish(exchange Exchange, route string, msg Message) error {
	produce, err := DefaultRMQ.RegisterProcedure(exchange)
	if err != nil {
		log.WithError(err).Error("生产者创建失败")
		return err
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.WithError(err).Error("消息序列化失败")
		return err
	}
	/*log.WithFields(logrus.Fields{
		"exchange":   exchange.Name,
		"routingKey": route,
		"messageID":  msg.MessageID,
	}).Debug("发送RabbitMq消息")*/
	produce.PublishSimple(route, bytes)
	return nil
}
