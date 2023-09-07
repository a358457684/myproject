package util

import (
	"bytes"
	"sync"
)

var(
	_bufferPool		sync.Pool
)

func GetBuffer()*bytes.Buffer  {
	v := _bufferPool.Get()
	if v == nil{
		return bytes.NewBuffer(make([]byte,0,256))
	}
	return v.(*bytes.Buffer)
}

func FreeBuffer(buffer *bytes.Buffer)  {
	buffer.Reset()
	_bufferPool.Put(buffer)
}
