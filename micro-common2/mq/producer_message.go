package mq

type ProducerMessage struct {
	Topic string
	Value interface{}
}
