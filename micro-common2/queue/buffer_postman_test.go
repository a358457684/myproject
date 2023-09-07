package queue

import (
	"fmt"
	"pp/common-golang/utils"
	"testing"
	"time"
)

var (
	testQueue chan interface{}
)

func init() {
	testQueue = make(chan interface{}, 500)
}

type TestMessage struct {
	ID    utils.Long
	Count int
}

func (m *TestMessage) BufferID() interface{} {
	return m.ID
}

func (m *TestMessage) Reduce(oldVal IBufferItem) IBufferItem {
	return m // 此行示例即不管旧值，直接返回新值，Push后的效果即新值直接替换旧值
	// 或者实现累加之类的逻辑
	//m.Count = oldVal.(*TestMessage).Count + m.Count
	//return m
}

func TestBufferPostman_Push(t *testing.T) {
	// 初始化缓冲投递员
	bufPost := NewBufferPostman(5, 5*time.Second, testQueue)

	// 启动并发的消息消费者chan池
	NewChanDispatcher(testQueue, 4).Run(
		func(workerId int, msg interface{}) {
			if i, ok := msg.(IBufferItem); ok {
				fmt.Printf("consumer %d recevie message %d\n", workerId, i.BufferID())
			}
		},
	)

	// 准备一个测试去重用的id列表
	var ids []utils.Long
	for i := 0; i < 17; i++ {
		ids = append(ids, utils.NextId())
	}

	// 消息并发压进缓冲
	// 整个数据流：消息加进缓冲的map -> 超时或超限时，取出全部消息推进中转chan -> 中转chan消费者处理消息
	// 上述数据流入口，也可以是消费其它chan或消息中间件的消息，然后再加进map
	// 中转chan消费者不一定要用并发chan池，例如，也可以根据id散列，启动固定的几个消息消费者 -- 这对向数据库写数据更加合适，不易因为并发造成死锁
	for g := 0; g < 3; g++ {
		go func() {
			// 测试超时和超限机制
			for i := 0; i < 50; i++ {
				time.Sleep(100 * time.Millisecond)
				var id utils.Long = 0
				if nil != ids && len(ids) > 0 {
					id = ids[i%len(ids)]
				}
				bufPost.Push(&TestMessage{
					ID:    id,
					Count: i,
				})
			}
			// 当触发一次无数据时，超时定时器应该会停止
		}()
	}
	// 测试超时定时器唤醒机制
	go func() {
		time.Sleep(30 * time.Second)
		println(`30 seconds later`)
		bufPost.Push(&TestMessage{
			ID:    ids[0],
			Count: 0,
		})
		time.Sleep(10 * time.Second)
		for g := 0; g < 3; g++ {
			go func() {
				for i := 0; i < 500; i++ {
					time.Sleep(100 * time.Millisecond)
					var id utils.Long = 0
					if nil != ids && len(ids) > 0 {
						id = ids[i%len(ids)]
					}
					bufPost.Push(&TestMessage{
						ID:    id,
						Count: i,
					})
				}
			}()
		}
	}()

	quit := make(chan bool)
	<-quit
}
