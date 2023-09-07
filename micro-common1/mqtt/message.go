package mqtt

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Message struct {
	MessageID string
	ClientID  string
	Data      interface{}
	Time      time.Time
}

func NewMessage(data interface{}) Message {
	message := Message{
		MessageID: uuid.NewV4().String(),
		Data:      data,
		Time:      time.Now(),
	}
	return message
}

func (msg *Message) UnMarshal(v interface{}) error {
	bytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, v)
	if err != nil {
		return err
	}
	return nil
}
