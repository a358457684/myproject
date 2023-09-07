package mq

import "io"

type IProducer interface {
	io.Closer
	SendJSON(topic string, value interface{}) (interface{}, error) // 发布JSON消息
	SendJSONs(messages []*ProducerMessage) error                   // 批量发布JSON消息
}
