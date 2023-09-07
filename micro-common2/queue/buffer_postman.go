package queue

import (
	"pp/common-golang/utils"
	"sync"
	"time"
)

// 缓冲消息接口
type IBufferItem interface {
	BufferID() interface{}                 // 去重用的键，包括后续消费消息，如果需要也可以根据这个键做散列处理
	Reduce(oldVal IBufferItem) IBufferItem // 接口实现类中实现此方法，以实现累加之类的多态的业务逻辑，当然最简单不做其他处理直接返回新的对象值自身也行
}

// 缓冲键值表
type BufferMap struct {
	data map[interface{}]IBufferItem
	lock *sync.Mutex
}

// 新建缓冲键值表对象
func NewBufferMap() *BufferMap {
	return &BufferMap{
		data: make(map[interface{}]IBufferItem),
		lock: new(sync.Mutex),
	}
}

// 置入键值
func (m *BufferMap) Push(item IBufferItem) int {
	if nil == item {
		return 0
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	oldVal := m.data[item.BufferID()]
	m.data[item.BufferID()] = item.Reduce(oldVal)
	return len(m.data)
}

// 读取键值
func (m *BufferMap) Get(id interface{}) IBufferItem {
	if nil == id {
		return nil
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.data[id]
}

// 读取并移除键值
func (m *BufferMap) Pop(id interface{}) IBufferItem {
	if nil == id {
		return nil
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	res := m.data[id]
	delete(m.data, id)
	return res
}

// 读取全部键
func (m *BufferMap) Keys() []interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.data) == 0 {
		return nil
	}
	var res []interface{}
	for k := range m.data {
		res = append(res, k)
	}
	return res
}

// 读取全部值
func (m *BufferMap) Values() []IBufferItem {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.data) == 0 {
		return nil
	}
	var res []IBufferItem
	for _, v := range m.data {
		res = append(res, v)
	}
	return res
}

// 读取并移除全部键值
func (m *BufferMap) PopAll() []IBufferItem {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.data) == 0 {
		return nil
	}
	var res []IBufferItem
	for _, v := range m.data {
		res = append(res, v)
	}
	m.data = make(map[interface{}]IBufferItem)
	return res
}

// 获取大小
func (m *BufferMap) Size() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return len(m.data)
}

// 移除键值
func (m *BufferMap) Remove(id interface{}) {
	if nil == id {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.data) == 0 {
		return
	}
	delete(m.data, id)
}

// 清空键值表
func (m *BufferMap) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.data) == 0 {
		return
	}
	m.data = make(map[interface{}]IBufferItem)
}

// 缓冲投递员
type BufferPostman struct {
	limit       int
	duration    time.Duration
	Buffer      *BufferMap
	timer       *time.Timer
	isTimerStop bool
	target      chan interface{}
}

// 新建缓冲投递员对象
func NewBufferPostman(limit int, duration time.Duration, target chan interface{}) *BufferPostman {
	p := &BufferPostman{
		limit:    limit,
		duration: duration,
		Buffer:   NewBufferMap(),
		target:   target,
	}
	if duration > 0 {
		p.timer = time.NewTimer(duration)
		go func(p *BufferPostman) {
			defer utils.DefaultGoroutineRecover(nil, `缓冲投递超时消息`)
			for {
				select {
				case <-p.timer.C: // 超时
					p.isTimerStop = true
					p.deliver()
				}
			}
		}(p)
	}
	return p
}

// 置入消息
func (p *BufferPostman) Push(item IBufferItem) {
	size := p.Buffer.Push(item)
	if p.isTimerStop { // 唤醒定时器
		p.resetTimer()
	}
	if p.limit > 0 && size >= p.limit { // 超限
		p.deliver()
	}
}

// 投递消息
func (p *BufferPostman) deliver() {
	data := p.Buffer.PopAll()
	if nil != data && len(data) > 0 {
		// 仅供测试时用的日志记录
		//if eventType == 1 {
		//	fmt.Printf("deliver %d messages when timeout\n", len(data))
		//} else if eventType == 2 {
		//	fmt.Printf("deliver %d messages when full\n", len(data))
		//}

		// 将消息推进中转chan
		go func(p *BufferPostman, data []IBufferItem) {
			defer utils.DefaultGoroutineRecover(nil, `缓冲投递超限消息`)
			for i := 0; i < len(data); i++ {
				p.target <- data[i]
			}
		}(p, data)

		// 触发时有数据才会重置定时器
		p.resetTimer()
	}
	// 没数据，定时器就暂停了，等待新数据进入时再唤醒
}

// 重置定时器
func (p *BufferPostman) resetTimer() {
	if nil != p.timer && p.duration > 0 {
		if len(p.timer.C) > 0 {
			<-p.timer.C
		}
		p.timer.Reset(p.duration)
		p.isTimerStop = false
	}
}
