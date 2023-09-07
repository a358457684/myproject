package kafka

import (
	"fmt"
	"pp/common-golang/mq"
	"pp/common-golang/utils"
	"sync"
	"testing"
)

var wg sync.WaitGroup

func TestNewConsumer(t *testing.T) {
	// 以下示例代码是模拟中断延时处理定时器需求实现，逻辑较复杂，如果只关心kafka使用本身，仅关注加了序号注释的部位即可

	// 1. 新建单播消费者
	// 初始化增加定时任务消息消费者，相同group名单播消费消息
	// kafka的单播分配到哪一个消费者与topic的partition配置密切相关，当partition数小于消费者数量时，会有部分消费者始终无法获得消息
	// 可以用 ./kafka-topics.sh --zookeeper 127.0.0.1:2181 --describe --topic topic-name 命令查看topic信息
	// 可以用 ./kafka-topics.sh --alter --topic topic-name --zookeeper 127.0.0.1:2181 --partitions 4 命令调整partition数量
	// 调整完partition数量，需要重启或重置消费者，以重新分配消费者与partition的对应关系
	// 另外注意本示例中用的topic.Order方法是有给topic加上一个前缀的
	plusTopics := []string{"test.ka-alloc-plus-job"}
	plusConsumer := NewConsumer("ka-alloc-plus-job-group", plusTopics)
	defer func(plusConsumer *Consumer) {
		_ = plusConsumer.Close()
	}(plusConsumer)
	// 2. 如果需要可以处理异常
	go func() {
		for err := range plusConsumer.Errors() {
			t.Errorf("Error: %s\n", err.Error())
		}
	}()

	// 1. 新建广播消费者
	// 初始化减少定时任务消息消费者，不同group名广播消费消息
	reduceTopics := []string{"test.ka-alloc-reduce-job"}
	// 正式编码推荐用 utils.GetPrivateIPv4Id() 代替 utils.NextId() 来获取机器号
	// utils.GetPrivateIPv4Id() 这个函数会根据当前机器内网ip的末尾两段运算出一个id，即测试和生产环境不同Pod的ip不同这个id也会不同
	machineId := utils.NextId()
	reduceConsumer := NewConsumer(fmt.Sprintf("ka-alloc-reduce-job-group-%d", machineId), reduceTopics)
	defer func(reduceConsumer *Consumer) {
		_ = reduceConsumer.Close()
	}(reduceConsumer)
	// 2. 如果需要可以处理异常
	go func() {
		for err := range reduceConsumer.Errors() {
			t.Errorf("Error: %s\n", err.Error())
		}
	}()

	// 消费消息
	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	for msg := range plusConsumer.Messages() { // 3. 在协程中循环阻塞取消息管道中的消息
	//		fmt.Printf( "%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value) // msg.Value如果是Json可以反序列化
	//		plusConsumer.MarkOffset(msg, "") // mark message as processed
	//	}
	//}()
	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	for msg := range reduceConsumer.Messages() { // 3. 在协程中循环阻塞取消息管道中的消息
	//		fmt.Printf( "%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value) // msg.Value如果是Json可以反序列化
	//		reduceConsumer.MarkOffset(msg, "") // mark message as processed
	//	}
	//}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var plusCh = make(chan *mq.TestMsg, 0)
		plusConsumer.BindJSONChan(plusCh)
		for { // 3. 在协程中循环阻塞取消息管道中的消息
			select {
			case msg := <-plusCh:
				fmt.Printf("%v\n", msg) // msg已经是反序列化得到的对象
			}
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		var reduceCh = make(chan *mq.TestMsg, 0)
		reduceConsumer.BindJSONChan(reduceCh)
		for { // 3. 在协程中循环阻塞取消息管道中的消息
			select {
			case msg := <-reduceCh:
				fmt.Printf("%v\n", msg) // msg已经是反序列化得到的对象
			}
		}
	}()

	wg.Wait()
	t.Log("Done consuming topic")
}
