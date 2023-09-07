package mqtt

import (
	"common/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
	"testing"
)

type User struct {
	Name string
	Age  int
}

func TestMqttExported(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(3)

	err := Subscribe(func(c *Client, topic string, msg *Message) {
		log.Infof("接收到消息----：%v", msg)
		waitGroup.Done()
	}, 1, "proxy-test-mqtt")
	if err != nil {
		panic(err)
	}

	err = Subscribe(func(c *Client, topic string, msg *Message) {
		log.Infof("接收到消息=======：%v", msg)
		waitGroup.Done()
	}, 1, "proxy-test-emqtt2")
	if err != nil {
		panic(err)
	}

	err = Publish("proxy-test-mqtt", 1, false, NewMessage("南无阿弥陀佛"))

	//data := make(map[string]interface{})
	//data["name"] = "张三"
	//data["age"] = 14
	data := User{
		Name: "张三",
		Age:  17,
	}
	err = Publish("proxy-test-emqtt2", 1, false, NewMessage(data))
	if err != nil {
		panic(err)
	}
	waitGroup.Wait()
}

func TestMqttCustomExported(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	err := SubscribeCustom(func(c *Client, data interface{}, message mqtt.Message) {
		log.Infof("接收到消息----：%v", data.(string))
		waitGroup.Done()
	}, "", 1, "proxy-test-mqtt")
	if err != nil {
		panic(err)
	}

	err = SubscribeCustom(func(c *Client, data interface{}, message mqtt.Message) {
		log.Infof("接收到消息=======：%v", data.(*User))
		waitGroup.Done()
	}, User{}, 2, "proxy-test-emqtt2")
	if err != nil {
		panic(err)
	}

	err = PublishCustom("proxy-test-mqtt", 1, false, "南无阿弥陀佛")

	data := User{
		Name: "张三",
		Age:  17,
	}
	err = PublishCustom("proxy-test-emqtt2", 1, false, data)
	if err != nil {
		panic(err)
	}
	waitGroup.Wait()
}
