package cache

import (
	"common/biz/enum"
	"common/redis"
	"common/util"
	"context"
	"time"
)

const (
	msgRetryKey = "msg_retry"
)

type JobCancelMsg struct {
	JobId string
}

type MsgRetryVo struct {
	MsgID     string             //消息ID
	Status    enum.MsgStatusEnum //状态
	Type      enum.MsgTypeEnum   //消息类型
	Topic     string             //topic
	Data      interface{}        //数据
	SendTimes int                //发送次数
	SendTime  time.Time          //最后一次发送时间
	Time      time.Time          //创建时间
}

func SaveMsgRetry(data MsgRetryVo) error {
	return redis.HSetJson(context.Background(), msgRetryKey, data.MsgID, data)
}

func GetMsgRetry(msgId string) (MsgRetryVo, error) {
	data := MsgRetryVo{}
	err := redis.HGetJson(context.Background(), &data, msgRetryKey, msgId)
	return data, err
}

func FindMsgRetryAll() ([]MsgRetryVo, error) {
	var data []MsgRetryVo
	if err := redis.HGetALLJson(context.Background(), &data, msgRetryKey); err != nil {
		return nil, util.WrapErr(err, "查询消息缓存失败")
	}
	return data, nil
}
