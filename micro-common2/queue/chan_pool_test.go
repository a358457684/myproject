package queue

import (
	"fmt"
	"sync"
	"testing"
)

type testMsg struct {
	ID int
}

// "线程池"（golang管道池）单元测试
func TestChanPool(t *testing.T) {
	// 参数
	maxWorkers := 4  // 最大工作管道数
	msgCount := 1000 // 消息数量

	// 非主逻辑，计数器初始化
	wg := sync.WaitGroup{} // 同步计数器
	wg.Add(msgCount)
	counter := NewCountMap() // 并发安全统计计数器

	// 启动工作管道池调度器
	msgQueue := make(chan interface{}, 0) // 输入队列
	NewChanDispatcher(msgQueue, maxWorkers).Run(
		//NewChanDispatcherWithCapacity(msgQueue, maxWorkers, 5).Run(
		func(workerId int, msg interface{}) {
			if n, ok := msg.(*testMsg); ok {
				fmt.Printf("worker %d received msg %d\n", workerId, n.ID)
				counter.Set(workerId, counter.Get(workerId)+1)
				wg.Done()
			}
		},
	)

	// 发消息给输入队列
	for i := 0; i < msgCount; i++ {
		msgQueue <- &testMsg{ID: i}
	}

	// 非主逻辑，计数器打印
	wg.Wait()
	println("")
	for k, v := range counter.Data {
		fmt.Printf("worker %d received msg total count is %d\n", k, v)
	}
}

type countMap struct {
	Data map[int]int
	Lock sync.Mutex
}

func NewCountMap() *countMap {
	return &countMap{
		Data: make(map[int]int),
	}
}

func (d *countMap) Get(k int) int {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	return d.Data[k]
}

func (d *countMap) Set(k, v int) {
	d.Lock.Lock()
	defer d.Lock.Unlock()
	d.Data[k] = v
}
