package mqtt

import (
	"common/config"
	"common/log"
	"encoding/json"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"reflect"
)

var uninitializedErr = errors.New("mqtt is uninitialized")
var (
	StringType    = reflect.TypeOf((*string)(nil)).Elem()
	ByteArrayType = reflect.TypeOf((*[]byte)(nil)).Elem()
)

func init() {
	if config.Data.Mqtt == nil {
		log.Warn("读取MQTT配置失败, 跳过MQTT初始化")
		return
	}
	err := initMQTT(config.Data.Mqtt)
	if err != nil {
		log.WithError(err).Error("MQTT初始化失败")
		panic(err)
	} else {
		log.Info("MQTT初始化成功")
	}
}

func GetClientId() string {
	if !IsConnected() {
		return ""
	}
	return client.getClientId()
}

func IsConnected() bool {
	return client != nil && client.IsConnected()
}

func Publish(topic string, qos byte, retained bool, message Message) error {
	message.ClientID = client.getClientId()
	return PublishCustom(topic, qos, retained, message)
}

func PublishCustom(topic string, qos byte, retained bool, data interface{}) error {
	if !IsConnected() {
		log.Warn(uninitializedErr)
		return uninitializedErr
	}
	return client.Publish(topic, qos, retained, data)
}

func Subscribe(observer func(c *Client, topic string, msg *Message), qos byte, topics ...string) error {
	return SubscribeCustom(func(c *Client, data interface{}, message mqtt.Message) {
		observer(c, message.Topic(), data.(*Message))
	}, Message{}, qos, topics...)
}

func SubscribeCustom(handler func(*Client, interface{}, mqtt.Message), msgType interface{}, qos byte, topics ...string) error {
	if !IsConnected() {
		return uninitializedErr
	}
	var t reflect.Type
	if msgType != nil {
		t = reflect.TypeOf(msgType)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	messageHandler := func(c mqtt.Client, msg mqtt.Message) {
		log.WithFields(logrus.Fields{
			"topic": msg.Topic(),
		}).Debug("Received Mqtt message")
		var data interface{}
		if t != nil {
			switch t {
			case ByteArrayType:
				data = msg.Payload()
			case StringType:
				data = string(msg.Payload())
			default:
				data = reflect.New(t).Interface()
				err := json.Unmarshal(msg.Payload(), data)
				if err != nil {
					log.LogWithError(err, "Failed to decode mqtt message", msg.Payload())
					return
				}
			}
		}
		handler(client, data, msg)
	}
	return client.Subscribe(messageHandler, qos, topics...)
}

func Unsubscribe(topics ...string) error {
	if !IsConnected() {
		return uninitializedErr
	}
	return client.Unsubscribe(topics...)
}
