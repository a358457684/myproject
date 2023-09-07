package mqtt

import (
	"common/config"
	"common/log"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

var client *Client

type Client struct {
	nativeClient  mqtt.Client
	clientOptions *mqtt.ClientOptions
	locker        *sync.Mutex
	topicMap      map[string]byte
	handlerMap    map[string]func(c mqtt.Client, msg mqtt.Message)
}

func initMQTT(options *config.MqttOptions) error {
	log.Info("开始初始化MQTT...")
	clientOptions := mqtt.NewClientOptions()
	for _, service := range options.Services {
		if !strings.Contains(service, "tcp") {
			service = fmt.Sprintf("tcp://%s", service)
		}
		clientOptions.AddBroker(service)
	}
	clientOptions.Username = options.Username
	clientOptions.Password = options.Password
	clientOptions.ClientID = options.ClientId + ":" + uuid.NewV4().String()

	clientOptions.OnConnect = defaultOnConnectHandler
	clientOptions.OnConnectionLost = defaultConnectionLostHandler

	myLog := log.WithFields(logrus.Fields{
		"service":  clientOptions.Servers,
		"clientId": clientOptions.ClientID,
	})
	myLog.Info("开始连接MQTT服务...")

	// 创建客户端连接
	nativeClient := mqtt.NewClient(clientOptions)

	client = &Client{
		nativeClient:  nativeClient,
		clientOptions: clientOptions,
		locker:        &sync.Mutex{},
		topicMap:      make(map[string]byte),
		handlerMap:    make(map[string]func(c mqtt.Client, msg mqtt.Message)),
	}

	err := client.Connect()
	return err
}

func (client *Client) getClientId() string {
	return client.clientOptions.ClientID
}

func (client *Client) IsConnected() bool {
	return client.nativeClient.IsConnected()
}

// 连接
func (client *Client) Connect() error {
	return client.ensureConnected()
}

// 确保连接
func (client *Client) ensureConnected() error {
	if !client.nativeClient.IsConnected() {
		client.locker.Lock()
		defer client.locker.Unlock()
		if !client.nativeClient.IsConnected() {
			if token := client.nativeClient.Connect(); token.Wait() && token.Error() != nil {
				return token.Error()
			}
		}
	}
	return nil
}

// 发送消息
func (client *Client) Publish(topic string, qos byte, retained bool, data interface{}) error {
	if err := client.ensureConnected(); err != nil {
		return err
	}
	log.WithFields(logrus.Fields{
		"topic":    topic,
		"qos":      qos,
		"retained": retained,
	}).Debug("Send Mqtt message")
	var bytes []byte
	switch vData := data.(type) {
	case []byte:
		bytes = vData
	case *[]byte:
		bytes = *vData
	case string:
		bytes = []byte(vData)
	case *string:
		bytes = []byte(*vData)
	default:
		var err error
		bytes, err = json.Marshal(data)
		if err != nil {
			log.WithError(err).Error("消息序列化编码失败")
			return err
		}
	}
	token := client.nativeClient.Publish(topic, qos, retained, bytes)
	if err := token.Error(); err != nil {
		log.WithField("message", data).WithError(err).Error("Mqtt message publish failed!")
		return err
	}

	if !token.WaitTimeout(time.Second * 10) {
		log.WithField("message", data).Error("Mqtt message publish timeout!")
		return errors.New("mqtt publish wait timeout")
	}

	return nil
}

// 消费订阅
func (client *Client) Subscribe(messageHandler func(c mqtt.Client, msg mqtt.Message), qos byte, topics ...string) error {
	if err := client.ensureConnected(); err != nil {
		return err
	}
	if len(topics) == 0 {
		return errors.New("the topic is empty")
	}

	if messageHandler == nil {
		return errors.New("the observer func is nil")
	}

	filters := make(map[string]byte)
	for _, topic := range topics {
		filters[topic] = qos
	}

	token := client.nativeClient.SubscribeMultiple(filters, messageHandler)
	if token.Wait() && token.Error() != nil {
		log.WithFields(logrus.Fields{
			"topic":    topics,
			"clientId": client.getClientId(),
		}).WithError(token.Error()).Error("Failed to subscribe topic")
		return token.Error()
	}
	for _, topic := range topics {
		client.handlerMap[topic] = messageHandler
		client.topicMap[topic] = qos
	}
	log.WithFields(logrus.Fields{
		"topics": topics,
		"qos":    qos,
	}).Info("成功添加订阅")
	return nil
}

// 解绑订阅
func (client *Client) Unsubscribe(topics ...string) error {
	if err := client.ensureConnected(); err != nil {
		return err
	}
	token := client.nativeClient.Unsubscribe(topics...)
	if token.Wait() && token.Error() != nil {
		log.WithFields(logrus.Fields{
			"topic":    topics,
			"clientId": client.getClientId(),
		}).WithError(token.Error()).Error("Failed to unsubscribe topic")
		return token.Error()
	}
	return nil
}
