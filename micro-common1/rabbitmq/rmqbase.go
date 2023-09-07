package rabbitmq

import (
	"common/log"
	"github.com/streadway/amqp"
	"github.com/suiyunonghen/DxCommonLib"
	"sync/atomic"
	"time"
)

type ExchangeType uint8

const (
	ET_None ExchangeType = iota //不使用exchange，直接使用队列
	ET_Direct
	ET_Topic
	ET_Headers
	ET_Fanout
)

func (t ExchangeType) String() string {
	switch t {
	case ET_Direct:
		return "direct"
	case ET_Fanout:
		return "fanout"
	case ET_Topic:
		return "topic"
	case ET_Headers:
		return "headers"
	case ET_None:
		return ""
	}
	return "fanout"
}

type RQStatus int32

const (
	RQSOffline RQStatus = iota
	RQSReconnecting
	RQSActive
	RQSStart
)

type Exchange struct {
	Model ExchangeType
	Name  string
}

type RMQBase struct {
	state           RQStatus //状态
	rqurl           string
	l               *log.Logger
	rmqCon          *amqp.Connection
	channel         *amqp.Channel
	rmqConnectOK    func() error
	beforeReconnect func() //重连之前执行
	quit            chan struct{}
}

func (p *RMQBase) Close() {
	state := RQStatus(atomic.LoadInt32((*int32)(&p.state)))
	if state == RQSOffline {
		return
	}
	close(p.quit)
	if state == RQSActive {
		p.channel.Close()
		p.rmqCon.Close()
		atomic.StoreInt32((*int32)(&p.state), int32(RQSOffline))
	}
}

func (p *RMQBase) Status() RQStatus {
	return RQStatus(atomic.LoadInt32((*int32)(&p.state)))
}

func (p *RMQBase) reconnect(data ...interface{}) {
	if p.beforeReconnect != nil {
		p.beforeReconnect()
	}
	if !p.rmqCon.IsClosed() {
		p.rmqCon.Close()
	}
	//执行重连
	idx := 0
	for {
		idx++
		if idx > 10 {
			idx = 1
		}
		con, err := amqp.Dial(p.rqurl)
		if err != nil {
			if p.l != nil {
				p.l.WithError(err).Error("重连失败")
			}
			DxCommonLib.Sleep(time.Second * time.Duration(idx))
			continue
		}
		c, err := con.Channel()
		if err != nil {
			if p.l != nil {
				p.l.WithError(err).Error("RMQ重连失败")
			}
			con.Close()
			DxCommonLib.Sleep(time.Second * time.Duration(idx))
			continue
		}
		p.rmqCon = con
		p.channel = c
		if p.rmqConnectOK != nil {
			if err = p.rmqConnectOK(); err != nil {
				DxCommonLib.Sleep(time.Second * time.Duration(idx))
				continue
			}
		}
		atomic.StoreInt32((*int32)(&p.state), int32(RQSActive))
		if p.l != nil {
			p.l.Debugf("RMQ重连成功,Address=%s", p.rqurl)
		}
		//重连OK
		p.enableReconnect(true)
		return
	}

}

//启动重连机制
func (p *RMQBase) enableReconnect(autoReconnect bool) {
	if p.rmqCon == nil || p.channel == nil {
		return
	}
	conerr := make(chan *amqp.Error)
	channelErr := make(chan *amqp.Error)
	p.rmqCon.NotifyClose(conerr)
	p.channel.NotifyClose(channelErr)

	checkifReconnect := func(msginfo string, err *amqp.Error) {
		if autoReconnect {
			atomic.StoreInt32((*int32)(&p.state), int32(RQSReconnecting))
			DxCommonLib.MustRunAsync(p.reconnect)
			if p.l != nil {
				p.l.WithError(err).Error(msginfo)
			}
		} else {
			if p.beforeReconnect != nil {
				p.beforeReconnect()
			}
			atomic.StoreInt32((*int32)(&p.state), int32(RQSOffline))
			if !p.rmqCon.IsClosed() {
				p.rmqCon.Close()
			}
		}
	}
	go func() {
		for {
			select {
			case err := <-conerr:
				//连接断开，准备执行重连机制
				checkifReconnect("RabbitMQ连接断开，准备自动重连", err)

				return
			case err := <-channelErr:
				//准备重连
				checkifReconnect("RabbitMQ的channel断开，准备自动重连", err)
				return
			case <-p.quit:
				atomic.StoreInt32((*int32)(&p.state), int32(RQSOffline))
				return
			}
		}
	}()
}
