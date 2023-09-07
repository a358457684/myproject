package rabbitmq

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Message struct {
	MessageID string
	Data      interface{}
	Time      time.Time
}

func NewMessage(data interface{}) Message {
	return Message{
		MessageID: uuid.NewV4().String(),
		Data:      data,
		Time:      time.Now(),
	}
}

func (msg *Message) UnMarshal(v interface{}) error {
	fmt.Println("data:",msg.Data)
	bytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}
	fmt.Println("bytes:",string(bytes))
	err = json.Unmarshal(bytes, v)
	if err != nil {
		return err
	}
	return nil
}
