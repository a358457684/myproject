package rabbitmq

import (
	"common/config"
	"common/log"
	"common/util"
	"errors"
	"github.com/streadway/amqp"
	"github.com/suiyunonghen/DxCommonLib"
	"github.com/suiyunonghen/dxsvalue"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

//生产者
type RbProcedure struct {
	Exchange
	ownerRMQ *RabbitMQ //所属的RabbitMQ
}

func (procedure *RbProcedure) PublishSimple(routkey string, msg []byte) error {
	if procedure.ownerRMQ == nil {
		return errors.New("没有绑定到RabbitMQ")
	}
	switch RQStatus(atomic.LoadInt32((*int32)(&procedure.ownerRMQ.state))) {
	case RQSActive:
		//发送
		if procedure.Model == ET_None {
			return procedure.ownerRMQ.channel.Publish("", procedure.Name, false, false,
				amqp.Publishing{Body: msg,
					DeliveryMode: 2, //2表示持久化，
					ContentType:  "application/json"})
		} else {
			return procedure.ownerRMQ.channel.Publish(procedure.Name, routkey, false, false,
				amqp.Publishing{Body: msg,
					DeliveryMode: 2, //2表示持久化，
					ContentType:  "application/json"})
		}
	case RQSOffline:
		return io.EOF
	case RQSReconnecting:
		//正在重连中将消息记录到缓存先
		return errors.New("已经断线，正在重连中")
	}
	return nil
}

func (procedure *RbProcedure) Publish(routkey string, publishing amqp.Publishing) error {
	if procedure.ownerRMQ == nil {
		return errors.New("没有绑定到RabbitMQ")
	}
	switch RQStatus(atomic.LoadInt32((*int32)(&procedure.ownerRMQ.state))) {
	case RQSActive:
		//发送
		if procedure.Model == ET_None {
			return procedure.ownerRMQ.channel.Publish("", procedure.Name, false, false, publishing)
		} else {
			return procedure.ownerRMQ.channel.Publish(procedure.Name, routkey, false, false, publishing)
		}
	case RQSOffline:
		return io.EOF
	case RQSReconnecting:
		//正在重连中将消息记录到缓存先
		return errors.New("已经断线，正在重连中")
	}
	return nil
}

func (procedure *RbProcedure) reConnect() {
	//重连
	if procedure.Model == ET_None {
		if _, err := procedure.ownerRMQ.channel.QueueDeclare(procedure.Name,
			true, true, false, true, nil); err != nil {
			if procedure.ownerRMQ.l != nil {
				procedure.ownerRMQ.l.WithError(err).WithField("queue", procedure.Name).Error("RMQ生产者重建队列错误")
			}
		}
	} else if err := procedure.ownerRMQ.channel.ExchangeDeclare(procedure.Name, procedure.Model.String(),
		true, false, false, true, nil); err != nil {
		if procedure.ownerRMQ.l != nil {
			procedure.ownerRMQ.l.WithError(err).WithField("exchange", procedure.Name).Error("RMQ生产者重建发生错误")
		}
	}
}

type RbConsume struct {
	autoAck bool
	Exchange
	ownerRMQ  *RabbitMQ
	QueueName string //队列名称
	bindings  []string
	handler   func(msg amqp.Delivery)
	sync.RWMutex
}

//订阅主题
func (c *RbConsume) Subscribe(topickey string) error {
	if c.Exchange.Name == "" || c.ownerRMQ.Status() != RQSActive {
		return nil
	}
	if topickey == "" && c.Exchange.Model == ET_Fanout {
		return nil
	}

	c.Lock()
	for _, vstr := range c.bindings {
		if vstr == topickey {
			c.Unlock()
			return nil
		}
	}
	if err := c.ownerRMQ.channel.QueueBind(c.QueueName, topickey, c.Exchange.Name, true, nil); err != nil {
		c.Unlock()
		return err
	}
	c.bindings = append(c.bindings, topickey)
	c.Unlock()
	return nil
}

func (c *RbConsume) reconnect() {
	//先注册
	l := c.ownerRMQ.l
	if c.Model != ET_None {
		if err := c.ownerRMQ.channel.ExchangeDeclare(c.Name, c.Model.String(),
			true, false, false, true, nil); err != nil {
			if l != nil {
				l.WithError(err).WithField("exchange", c.Name).Error("RMQ生产者重建Exchange错误")
			}
			return
		}
	}
	c.RLock()
	if _, err := c.ownerRMQ.channel.QueueDeclare(c.QueueName,
		true, true, false, true, nil); err != nil {
		c.RUnlock()
		if l != nil {
			l.WithError(err).WithField("queueName", c.QueueName).Error("RMQ生产者重建Queue错误")
		}
		return
	}
	if c.Model != ET_None {
		for i := 0; i < len(c.bindings); i++ {
			for _, bindkey := range c.bindings {
				if err := c.ownerRMQ.channel.QueueBind(c.QueueName, bindkey, c.Name, true, nil); err != nil {
					if l != nil {
						l.WithError(err).WithField("bindkey", bindkey).Error("RMQ生产者绑定queue错误")
					}
					continue
				}
			}
		}
	}
	c.RUnlock()
	msgs, err := c.ownerRMQ.channel.Consume(c.QueueName, "", c.autoAck, false, false, true, nil)
	if err != nil {
		if l != nil {
			l.WithError(err).Error("RMQ创建消费者错误")
		}
		return
	}
	c.Lock()
	willStartMonitor := c.ownerRMQ.consumeMsgs == nil
	if willStartMonitor {
		c.ownerRMQ.consumeMsgs = make(chan consumeMsg, 1)
	}
	c.Unlock()
	if willStartMonitor {
		go c.ownerRMQ.handleConsumeMsg(c.ownerRMQ.consumeMsgs)
	}
	c.ownerRMQ.consumeMsgs <- consumeMsg{
		consume: c,
		msg:     msgs,
	}
}

type consumeMsg struct {
	consume *RbConsume
	msg     <-chan amqp.Delivery
}

//将消费者和生产者公用一个连接
type RabbitMQ struct {
	RMQBase
	procedures            []RbProcedure
	consumes              []RbConsume
	consumeMsgs           chan consumeMsg
	DefaultConsumeHandler func(consume *RbConsume, msg amqp.Delivery)
	sync.RWMutex
}

var (
	uniqueueName sync.Map //exchange->queueName订阅有用
)

func saveUniqueueName(data ...interface{}) {
	json := dxsvalue.NewCacheValue(dxsvalue.VT_Object)
	cache := json.ValueCache()
	uniqueueName.Range(func(key, value interface{}) bool {
		json.SetKeyCached(key.(string), dxsvalue.VT_String, cache).SetString(value.(string))
		return true
	})
	if json.Count() > 0 {
		dxsvalue.Value2File(json, "./rmqQueue.json", false, true)
	}
	dxsvalue.FreeValue(json)
}

//注册一个生产者
func (rmq *RabbitMQ) RegisterProcedure(exchange Exchange) (*RbProcedure, error) {
	var result *RbProcedure
	rmq.Lock()
	for i := 0; i < len(rmq.procedures); i++ {
		if rmq.procedures[i].Model == exchange.Model &&
			rmq.procedures[i].Name == exchange.Name {
			//已经存在
			result = &rmq.procedures[i]
			break
		}
	}
	if result == nil {
		if exchange.Model == ET_None {
			if _, err := rmq.channel.QueueDeclare(exchange.Name, true, true, false, true, nil); err != nil {
				rmq.Unlock()
				return nil, err
			}
		} else if err := rmq.channel.ExchangeDeclare(exchange.Name, exchange.Model.String(), true, false, false, true, nil); err != nil {
			rmq.Unlock()
			return nil, err
		}
		rmq.procedures = append(rmq.procedures, RbProcedure{
			Exchange: exchange,
			ownerRMQ: rmq,
		})
		result = &rmq.procedures[len(rmq.procedures)-1]
	}
	rmq.Unlock()
	return result, nil
}

func (rmq *RabbitMQ) Publish(exchange Exchange, route string, msg []byte) error {
	proc, err := rmq.RegisterProcedure(exchange)
	if err != nil {
		return err
	}
	proc.PublishSimple(route, msg)
	return nil
}

func (rmq *RabbitMQ) reconnectOK() error {
	//重新注册之前已经存入的生产者
	rmq.Lock()
	for i := 0; i < len(rmq.procedures); i++ {
		rmq.procedures[i].reConnect()
	}

	for i := 0; i < len(rmq.consumes); i++ {
		rmq.consumes[i].reconnect()
	}
	rmq.Unlock()
	return nil
}

func (rmq *RabbitMQ) RegisterConsume(exchange Exchange, queueName string, autoAck bool, h func(d amqp.Delivery)) (*RbConsume, error) {
	if exchange.Model == ET_None {
		if queueName == "" {
			queueName = exchange.Name
		}
		exchange.Name = ""
	}
	if exchange.Name == "" && queueName == "" ||
		rmq.DefaultConsumeHandler == nil && h == nil {
		return nil, errors.New("无效的参数")
	}
	var result *RbConsume
	rmq.Lock()
	for i := 0; i < len(rmq.consumes); i++ {
		if rmq.consumes[i].Model == exchange.Model && rmq.consumes[i].Name == exchange.Name &&
			rmq.consumes[i].QueueName == queueName {
			result = &rmq.consumes[i]
			rmq.Unlock()
			return result, nil
		}
	}
	//先注册
	if exchange.Name != "" {
		if err := rmq.channel.ExchangeDeclare(exchange.Name, exchange.Model.String(), true, false, false, true, nil); err != nil {
			rmq.Unlock()
			return nil, err
		}
	}

	durable := true
	//autoDelete := false
	if queueName == "" {
		durable = false
		//autoDelete = true
		//不知道为啥设置了autoDelete还是一样会无畏的增加一系列的队列，所以，这里先查找本地配置，看看，有没有产生一个唯一的ID，如果有，就用这个唯一的ID来生成queueName
		if v, ok := uniqueueName.Load(exchange.Name); ok {
			queueName = v.(string)
		} else {
			//生成一个新的唯一ID
			queueName = config.Data.Project.Name + "_" + strconv.FormatInt(util.GetSnowflakeID(), 16)
			uniqueueName.Store(exchange.Name, queueName)
			DxCommonLib.MustRunAsync(saveUniqueueName)
		}
	}
	if queue, err := rmq.channel.QueueDeclare(queueName, durable, true, false, true, nil); err != nil {
		rmq.Unlock()
		return nil, err
	} else {
		queueName = queue.Name
	}
	//创建消费者
	msgs, err := rmq.channel.Consume(queueName, "", autoAck, false, false, true, nil)
	if err != nil {
		if rmq.l != nil {
			rmq.l.WithError(err).Error("RMQ创建消费者错误")
		}
		rmq.Unlock()
		return nil, err
	}

	rmq.consumes = append(rmq.consumes, RbConsume{
		Exchange:  exchange,
		handler:   h,
		QueueName: queueName,
		bindings:  make([]string, 0, 4),
		autoAck:   autoAck,
		ownerRMQ:  rmq,
	})
	result = &rmq.consumes[len(rmq.consumes)-1]
	willStartMonitor := rmq.consumeMsgs == nil
	if willStartMonitor {
		rmq.consumeMsgs = make(chan consumeMsg, 1)
	}
	rmq.Unlock()
	if willStartMonitor {
		go rmq.handleConsumeMsg(rmq.consumeMsgs)
	}
	rmq.consumeMsgs <- consumeMsg{
		consume: result,
		msg:     msgs,
	}
	return result, nil
}

//执行消费者消息监控
func (rmq *RabbitMQ) handleConsumeMsg(consumeChan <-chan consumeMsg) {
	tk := DxCommonLib.After(time.Second)
	consumers := make([]consumeMsg, 0, 8)
	quit := rmq.quit
	var wg sync.WaitGroup
	for {
		select {
		case <-quit:
			l := len(consumers)
			wg.Add(l)
			for i := 0; i < len(consumers); i++ {
				DxCommonLib.PostFunc(rmq.checkConsumerMsg, consumers[i], &wg)
			}
			wg.Done()
			rmq.Lock()
			close(rmq.consumeMsgs)
			rmq.consumeMsgs = nil
			rmq.Unlock()
			return
		case consumemsg, ok := <-consumeChan:
			if !ok {
				if rmq.l != nil {
					rmq.l.Debug("rabbitMQ断线了，退出连接")
				}
				return
			}
			consumers = append(consumers, consumemsg)
		case <-tk:
			if RQStatus(atomic.LoadInt32((*int32)(&rmq.state))) < RQSActive {
				return
			}
			tk = DxCommonLib.After(time.Second)
		default:
			//检查消费者的消息接收情况
			l := len(consumers)
			wg.Add(l)
			for i := 0; i < len(consumers); i++ {
				DxCommonLib.MustRunAsync(rmq.checkConsumerMsg, consumers[i], &wg)
			}
			wg.Wait()
		}
	}
}

func (rmq *RabbitMQ) checkConsumerMsg(data ...interface{}) {
	consumemsg := data[0].(consumeMsg)
	wg := data[1].(*sync.WaitGroup)
	select {
	case msg := <-consumemsg.msg:
		if msg.Body != nil {
			if consumemsg.consume.handler != nil {
				consumemsg.consume.handler(msg)
			} else if rmq.DefaultConsumeHandler != nil {
				rmq.DefaultConsumeHandler(consumemsg.consume, msg)
			}
		}
	case <-DxCommonLib.After(time.Millisecond * 500):
		//等待一下，超时退出
	}
	wg.Done()
}

func (rmq *RabbitMQ) dobeforeRecon() {
	rmq.Lock()
	if rmq.consumeMsgs != nil {
		close(rmq.consumeMsgs)
		rmq.consumeMsgs = nil
	}
	rmq.Unlock()
}

//构建一个RabbitMQ
func NewRabbitMQ(rmqAddr string, autoReconnect bool, l *log.Logger) (*RabbitMQ, error) {
	con, err := amqp.Dial(rmqAddr)
	if err != nil {
		return nil, err
	}
	c, err := con.Channel()
	if err != nil {
		return nil, err
	}
	rmq := &RabbitMQ{
		RMQBase: RMQBase{
			rqurl:   rmqAddr,
			l:       l,
			rmqCon:  con,
			channel: c,
			state:   RQSActive,
			quit:    make(chan struct{}),
		},
	}
	queueValue, err := dxsvalue.NewValueFromJsonFile("./rmqQueue.json", true)
	if err == nil {
		queueValue.Visit(func(Key string, value *dxsvalue.DxValue) bool {
			uniqueueName.Store(Key, value.String())
			return true
		})
	}
	//重连OK
	rmq.rmqConnectOK = rmq.reconnectOK
	rmq.beforeReconnect = rmq.dobeforeRecon
	rmq.enableReconnect(autoReconnect)
	return rmq, nil
}
