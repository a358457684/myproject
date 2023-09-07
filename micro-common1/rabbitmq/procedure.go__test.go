package rabbitmq

import (
	"common/config"
	"common/log"
	"fmt"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func TestExported(t *testing.T) {
	exchange := Exchange{
		Name:  "testCommonExchange",
		Model: ET_Topic,
	}

	_ = Subscribe(exchange, "test.#", "testCommonQueue", false, func(message *Message, delivery amqp.Delivery) {
		log.Infof("收到消息啦，%v", message)
		var data string
		_ = message.UnMarshal(&data)
		log.Info(data)
		// 手动确认消息
		_ = delivery.Ack(false)
	})
	_ = Publish(exchange, "test.one", NewMessage("南无阿弥陀佛"))
	time.Sleep(time.Second * 5)
}

func TestRabbitMQ_Publish(t *testing.T) {
	l := log.NewLog(&config.LogOptions{
		ShowCaller:  true,
		ShowConsole: true,
		JsonFormat:  true,
	})
	rmq, err := NewRabbitMQ("amqp://logistics-test:logistics-test@192.168.2.11:5672//logistics-test", true, l)
	if err != nil {
		fmt.Println("创建MQTT失败", err)
		return
	}
	exchange := Exchange{
		Name:  "dxsoft",
		Model: ET_Topic,
	}
	consume, err := rmq.RegisterConsume(exchange, "que3", true, func(d amqp.Delivery) {
		fmt.Println("test", d.RoutingKey, "消费者一收到消息", string(d.Body))
	})
	if err != nil {
		fmt.Println("订阅RMQ失败", err)
		return
	}
	consume.Subscribe("test.#") //订阅

	proc, err := rmq.RegisterProcedure(exchange)
	if err != nil {
		fmt.Println("创建Procedure失败", err)
		return
	}
	proc.PublishSimple("test.test", []byte(`{"ID":3}`))
	time.Sleep(time.Second * 10)
	fmt.Println("继续开始")
	proc.PublishSimple("test.123", []byte(`{"ID":5}`))
	time.Sleep(time.Second * 10)
}

func TestRabbitMQ_Publish2(t *testing.T) {
	exchange := Exchange{
		Model: ET_None,
		Name:  "队列Q",
	}
	Subscribe(exchange, "", "", true, func(message *Message, delivery amqp.Delivery) {
		fmt.Println("接收到消息:", message.Data)
	})
	err := Publish(exchange, "", NewMessage("测试内容"))
	if err != nil {
		fmt.Println("不建立exchange失败")
		return
	}
	time.Sleep(time.Minute)
}
