package kafka

import (
	"fmt"
	"github.com/CHneger/sarama-cluster"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
	"pp/common-golang/jsonutil"
	"pp/common-golang/logger"
	"pp/common-golang/utils"
	"reflect"
)

// 消费者对象
type Consumer struct {
	*cluster.Consumer
	log *logger.Logger
}

// 新建消费者
func NewConsumer(groupID string, topics []string) *Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = false

	viper.SetDefault("kafka.brokers", []string{"127.0.0.1:9092"})
	brokers := viper.GetStringSlice("kafka.brokers")

	consumer, err := cluster.NewConsumer(brokers, groupID, topics, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to start consumer: %s", err))
	}
	return &Consumer{Consumer: consumer}
}

// 将底层类库的日志输出到指定日志记录器
func (c *Consumer) SetLogger(log *logger.Logger) {
	if nil == log {
		return
	}
	c.log = log
	sarama.Logger = log
}

// 获取当前日志记录器
func (c *Consumer) GetLogger() *logger.Logger {
	if nil == c.log {
		c.log = logger.New()
	}
	return c.log
}

// 消息读取管道，管道消息类型是byte切片
func (c *Consumer) BytesMessages() <-chan []byte {
	ch := make(chan []byte, 0)
	go func(c *Consumer, ch chan []byte, oc <-chan *sarama.ConsumerMessage) {
		defer utils.DefaultGoroutineRecover(c.log, `KAFKA消息读取管道`)
		for msg := range oc {
			ch <- msg.Value
			c.MarkOffset(msg, "") // mark message as processed
		}
	}(c, ch, c.Consumer.Messages())
	return ch
}

// 将消息输出绑定到指定管道上，此方法内会进行反序列化，输出的消息类型可以是指定对象类型
func (c *Consumer) BindJSONChan(channel interface{}) {
	go func(c *Consumer, channel interface{}) {
		defer utils.DefaultGoroutineRecover(c.log, `KAFKA消息输出绑定`)
		chVal := reflect.ValueOf(channel)
		if chVal.Kind() != reflect.Chan {
			return
		}
		argType := chVal.Type().Elem()
		for {
			select {
			case msg := <-c.Messages():
				var oPtr reflect.Value
				if nil != msg && nil != msg.Value && len(msg.Value) > 0 && string(msg.Value) != "" {
					if argType.Kind() != reflect.Ptr {
						oPtr = reflect.New(argType)
					} else {
						oPtr = reflect.New(argType.Elem())
					}
					_ = jsonutil.Unmarshal(msg.Value, oPtr.Interface())
					if argType.Kind() != reflect.Ptr {
						oPtr = reflect.Indirect(oPtr)
					}
				}
				chVal.Send(oPtr)
				c.MarkOffset(msg, "") // mark message as processed
			}
		}
	}(c, channel)
}
