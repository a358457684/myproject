package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
	"pp/common-golang/jsonutil"
	"pp/common-golang/logger"
	"pp/common-golang/mq"
	"pp/common-golang/utils"
)

// kafka消息构造函数
func NewMsg(topic string, value interface{}) (msg *sarama.ProducerMessage, err error) {
	var bytes []byte
	if nil == value {
		bytes = []byte{}
	} else {
		bytes, err = jsonutil.Marshal(value)
		if nil != err {
			return
		}
	}
	msg = &sarama.ProducerMessage{
		Topic:     topic,
		Partition: int32(-1),                                     // 用于指定partition，仅当采用NewManualPartitioner时生效，但不同topic的partition数不一，手工指定很容易出现越界错误，一般不实用
		Key:       sarama.StringEncoder(utils.NextId().String()), // 当采用NewHashPartitioner时，是根据Key的hash值选取partition
		Value:     sarama.ByteEncoder(bytes),
	}
	return
}

// --------------------------------------------------------------------------------------------------------------------

// 生产者对象
type Producer struct {
	producer mq.IProducer
	log      *logger.Logger
}

// 新建生产者
func NewProducer() *Producer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 随机选取partition，还可以用NewRoundRobinPartitioner轮流选取，或者如前面的注释，可hash选取或手工指定
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3
	viper.SetDefault("kafka.brokers", []string{"127.0.0.1:9092"})
	brokers := viper.GetStringSlice("kafka.brokers")

	return &Producer{
		producer: NewSyncProducer(brokers, config), // 初始化同步或异步生产者
	}
}

// 将底层类库的日志输出到指定日志记录器
func (p *Producer) SetLogger(log *logger.Logger) {
	if nil == log {
		return
	}
	p.log = log
	sarama.Logger = log
}

// 获取当前日志记录器
func (p *Producer) GetLogger() *logger.Logger {
	if nil == p.log {
		p.log = logger.New()
	}
	return p.log
}

// 发送单条消息
func (p *Producer) SendJSON(topic string, value interface{}) (result interface{}, err error) {
	return p.producer.SendJSON(topic, value)
}

// 批量发送消息
func (p *Producer) SendJSONs(messages []*mq.ProducerMessage) (err error) {
	return p.producer.SendJSONs(messages)
}

// 关闭
func (p *Producer) Close() error {
	return p.producer.Close()
}

// --------------------------------------------------------------------------------------------------------------------

// 同步生产者
type SyncProducer struct {
	sarama.SyncProducer
}

// 新建同步生产者
func NewSyncProducer(brokers []string, config *sarama.Config) *SyncProducer {
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to produce message: %s", err))
	}
	return &SyncProducer{SyncProducer: producer}
}

// 同步生产者发送单条消息
func (p *SyncProducer) SendJSON(topic string, value interface{}) (result interface{}, err error) {
	if topic == "" {
		return "", nil
	}

	var msg *sarama.ProducerMessage
	msg, err = NewMsg(topic, value)
	if nil != err {
		return "", err
	}

	var partition int32
	var offset int64
	partition, offset, err = p.SendMessage(msg)
	if nil != err {
		return "", err
	}
	return fmt.Sprintf("partition=%d, offset=%d\n", partition, offset), nil
}

// 同步生产者批量发送消息
func (p *SyncProducer) SendJSONs(messages []*mq.ProducerMessage) (err error) {
	if nil == messages || len(messages) == 0 {
		return
	}
	var msgList []*sarama.ProducerMessage
	for i := 0; i < len(messages); i++ {
		if messages[i].Topic == "" {
			continue
		}

		var msg *sarama.ProducerMessage
		msg, err = NewMsg(messages[i].Topic, messages[i].Value)
		if nil != err {
			continue
		}

		msgList = append(msgList, msg)
	}

	return p.SendMessages(msgList)
}

// --------------------------------------------------------------------------------------------------------------------

// 异步生产者
type AsyncProducer struct {
	sarama.AsyncProducer
}

// 新建异步生产者
func NewAsyncProducer(brokers []string, config *sarama.Config) *AsyncProducer {
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to produce message: %s", err))
	}
	return &AsyncProducer{AsyncProducer: producer}
}

// 异步生产者发送单条消息
func (p *AsyncProducer) SendJSON(topic string, value interface{}) (result interface{}, err error) {
	if topic == "" {
		return "", nil
	}

	var msg *sarama.ProducerMessage
	msg, err = NewMsg(topic, value)
	if nil != err {
		return "", err
	}

	p.Input() <- msg
	return "success", nil
}

// 异步生产者批量发送消息
func (p *AsyncProducer) SendJSONs(messages []*mq.ProducerMessage) (err error) {
	if nil == messages || len(messages) == 0 {
		return
	}
	for i := 0; i < len(messages); i++ {
		if messages[i].Topic == "" {
			continue
		}

		var msg *sarama.ProducerMessage
		msg, err = NewMsg(messages[i].Topic, messages[i].Value)
		if nil != err {
			continue
		}

		p.Input() <- msg
	}

	return
}

// 注意异步生产者其实还有一个异步关闭的方法，且其需搭配下列结果处理代码使用
// 异步生消息生产者发送结果处理
//for {
//	select {
//	case suc := <-producer.Successes():
//		fmt.Println("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
//	case fail := <-producer.Errors():
//		fmt.Println("err: ", fail.Err)
//	}
//}
