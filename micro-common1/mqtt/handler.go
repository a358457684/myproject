package mqtt

import (
	"common/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

func defaultOnConnectHandler(c mqtt.Client) {
	reader := c.OptionsReader()
	log.WithFields(logrus.Fields{
		"service":  reader.Servers(),
		"clientId": reader.ClientID(),
	}).Info("Mqtt is connected!")
	for topic := range client.topicMap {
		handler := client.handlerMap[topic]
		qos := client.topicMap[topic]
		err := client.Subscribe(handler, qos, topic)
		if err != nil {
			log.WithFields(logrus.Fields{
				"service":  reader.Servers(),
				"clientId": reader.ClientID(),
				"topic":    topic,
			}).Info("mqtt恢复订阅失败")
		}
	}
}

func defaultConnectionLostHandler(client mqtt.Client, err error) {
	reader := client.OptionsReader()
	log.WithFields(logrus.Fields{
		"service":  reader.Servers(),
		"clientId": reader.ClientID(),
	}).WithError(err).Error("Mqtt is lostConnected! ")
}
