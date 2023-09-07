package mq

import "io"

type IConsumer interface {
	io.Closer
	BytesMessages() <-chan []byte     // 返回消息内容管道
	BindJSONChan(channel interface{}) // 绑定JSON管道，输出已经过反序列化的对象
}
