package cache

import (
	"common/biz/enum"
	"strings"
)

type Key struct {
	Data  string
	Items []string
}

func (k Key) Index(index int) string {
	if len(k.Items) == 0 {
		k.Items = strings.Split(k.Data, ":")
	}
	if len(k.Items) > index {
		return k.Items[index]
	}
	return ""
}

type OfficeRobotKey Key

func (t OfficeRobotKey) Category() string {
	return Key(t).Index(0)
}

func (t OfficeRobotKey) OfficeId() string {
	return Key(t).Index(1)
}

func (t OfficeRobotKey) RobotId() string {
	return Key(t).Index(2)
}

type MsgRetryKey Key

func (t MsgRetryKey) Category() string {
	return Key(t).Index(0)
}

func (t MsgRetryKey) MsgType() enum.MsgTypeEnum {
	return enum.MsgTypeEnum(Key(t).Index(1))
}

func (t MsgRetryKey) DestId() string {
	return Key(t).Index(2)
}

func (t MsgRetryKey) MsgId() string {
	return Key(t).Index(3)
}
