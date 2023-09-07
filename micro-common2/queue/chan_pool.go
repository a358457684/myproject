package queue

import (
	"os"
	"os/signal"
	"pp/common-golang/utils"
	"syscall"
)

// 工作对象封装
type ChanWorker struct {
	ID         int                   // 工作对象编号
	WorkerPool chan chan interface{} // 工作管道池，实例化时由调度器传入
	JobChannel chan interface{}      // 工作管道
	quit       chan bool             // 退出消息
}

func NewChanWorker(workerId int, workerPool chan chan interface{}) *ChanWorker {
	jobChannel := make(chan interface{})

	return &ChanWorker{
		ID:         workerId,
		WorkerPool: workerPool,
		JobChannel: jobChannel,
		quit:       make(chan bool),
	}
}

func (w *ChanWorker) Start(callback func(workerId int, msg interface{})) {
	go func(w *ChanWorker, callback func(workerId int, msg interface{})) {
		defer utils.DefaultGoroutineRecover(nil, `chan池工作对象消息处理`)
		for {
			// 新工作管道或每次取用工作管道后，加入工作管道池
			w.WorkerPool <- w.JobChannel

			select {
			case msg, ok := <-w.JobChannel: // 无消息时阻塞
				if ok {
					callback(w.ID, msg)
				}
			case <-w.quit:
				return
			}
		}
	}(w, callback)

	w.closeWait()
}

func (w *ChanWorker) closeWait() {
	go func(w *ChanWorker) {
		defer utils.DefaultGoroutineRecover(nil, `chan池关闭`)
		var c chan os.Signal
		var s os.Signal
		c = make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
		for {
			s = <-c
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				w.quit <- true
				return
			default:
				return
			}
		}
	}(w)
}

// 调度对象
type ChanDispatcher struct {
	MsgQueue   chan interface{}      // 消息输入管道
	WorkerPool chan chan interface{} // 工作管道池
	maxWorkers int                   // 最大工作对象数
}

func NewChanDispatcher(msgQueue chan interface{}, maxWorkers int) *ChanDispatcher {
	return &ChanDispatcher{
		MsgQueue:   msgQueue,
		WorkerPool: make(chan chan interface{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

func (d *ChanDispatcher) Run(callback func(workerId int, msg interface{})) {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewChanWorker(i, d.WorkerPool)
		worker.Start(callback)
	}

	d.dispatch()
}

func (d *ChanDispatcher) dispatch() {
	go func(d *ChanDispatcher) {
		defer utils.DefaultGoroutineRecover(nil, `chan池调度`)
		for {
			select {
			case msg, ok := <-d.MsgQueue:
				if ok {
					// 从工作管道池中尝试取出一个空闲的工作管道（每次取用工作管道会从池中取出去，消息处理完再放回池子，所以池子中的都是空闲的）
					// 无空闲工作管道（池子中无消息）时阻塞
					jobChannel, isOpen := <-d.WorkerPool
					if isOpen {
						// 将一条消息发送给成功取出的工作管道
						jobChannel <- msg
					}
				}
			}
		}
	}(d)

}
