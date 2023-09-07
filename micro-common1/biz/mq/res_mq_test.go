package mq

import (
	"common/log"
	"common/rabbitmq"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestResourceNotify_LiftLock(t *testing.T) {
	res := NewLiftResourceNotify(2, "@34234", "")
	res.LiftLock("gg", "sdfasdf", "@34234", 2, time.Now())
	res.LiftLock("test", "bb", "\"@34234", 2, time.Now())
	fmt.Println(res)
	message := res.createRabbitMQMsg()
	bt, _ := json.Marshal(message)
	fmt.Println(string(bt))

	var msg rabbitmq.Message
	var notifyData []ResourceNotify
	msg.Data = &notifyData
	err := json.Unmarshal(bt, &msg)
	if err != nil {
		log.WithError(err).Error("解析资源占用通知信息错误")
		return
	}
	fmt.Println(notifyData)
}
