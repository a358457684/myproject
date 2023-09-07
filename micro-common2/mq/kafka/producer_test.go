package kafka

import (
	"pp/common-golang/date"
	"pp/common-golang/mq"
	"testing"
)

func TestProducer_SendJSON(t *testing.T) {
	// 以下示例代码是模拟中断延时处理定时器需求实现，逻辑较复杂，如果只关心kafka使用本身，仅关注加了序号注释的部位即可

	// 1. 新建生产者
	producer := NewProducer()
	defer func(producer *Producer) {
		_ = producer.Close()
	}(producer)

	plusMsg := &mq.TestMsg{Message: "你好, 世界++!", Time: date.Now()}
	plusTopic := "test.ka-alloc-plus-job"
	reduceMsg := &mq.TestMsg{Message: "你好, 世界--!", Time: date.Now()}
	reduceTopic := "test.ka-alloc-reduce-job"
	var plusMsgList []*mq.ProducerMessage
	var reduceMsgList []*mq.ProducerMessage
	for i := 0; i < 5; i++ {
		plusMsgList = append(plusMsgList, &mq.ProducerMessage{
			Topic: plusTopic,
			Value: plusMsg,
		})
		result, err := producer.SendJSON(plusTopic, plusMsg) // 2. 发单条
		if err != nil {
			t.Error("Failed to produce message: ", err)
		}
		if nil != result {
			t.Log(result.(string))
		}

		reduceMsgList = append(reduceMsgList, &mq.ProducerMessage{
			Topic: reduceTopic,
			Value: reduceMsg,
		})
		result, err = producer.SendJSON(reduceTopic, reduceMsg) // 2. 发单条
		if err != nil {
			t.Error("Failed to produce message: ", err)
		}
		if nil != result {
			t.Log(result.(string))
		}
	}

	// 异步生消息生产者的发送结果处理
	//for i := 0; i < 5; i++ {
	//	select {
	//	case suc := <-producer.Successes():
	//		fmt.Println("offset: ", suc.Offset, "timestamp: ", suc.Timestamp.String(), "partitions: ", suc.Partition)
	//	case fail := <-producer.Errors():
	//		fmt.Println("err: ", fail.Err)
	//	}
	//}

	err := producer.SendJSONs(plusMsgList)
	if err != nil {
		t.Error("Failed to produce message: ", err) // 3. 发多条
	}

	err = producer.SendJSONs(reduceMsgList)
	if err != nil {
		t.Error("Failed to produce message: ", err) // 3. 发多条
	}
}
