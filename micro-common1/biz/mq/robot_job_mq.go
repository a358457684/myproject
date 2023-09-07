package mq

import (
	"common/biz/cache"
	"common/biz/enum"
	"common/log"
	"common/rabbitmq"
	"common/util"
	"encoding/json"
	"github.com/streadway/amqp"
	"time"
)

type robotJobRoute string

const (
	RjrStatus robotJobRoute = "status"
)

var (
	RobotJobStatusExchange = rabbitmq.Exchange{
		Name:  "job_status",
		Model: rabbitmq.ET_Direct,
	}
)

type BaseRobotJobInfo struct {
	AcceptState     enum.AcceptStatusEnum //物品接收状态 0 待接收；1已接收 ;2 超时未取
	JobInfo         cache.BaseJobInfo
	CompletedTime   time.Time `json:",omitempty"` //完成时间
	CompletedUserID string    //完成者
	Distance        int       //行驶里程
	Remarks         string    //备注，失败/取消原因
	OfficeID        string    //机构ID
}

type RobotJobStatus struct {
	BaseRobotJobInfo
	FailedTime  time.Time `json:"-"` //上次失败的时间，不序列化
	FirstFailed time.Time `json:"-"` //第一次失败的时间
}

var (
	_jobStatusProduce *rabbitmq.RbProcedure
	_jobStatusConsume *rabbitmq.RbConsume
)

//任务状态更新
func JobStatusUpdate(jobinf ...RobotJobStatus) error {
	if len(jobinf) == 0 {
		return nil
	}
	bf := util.GetBuffer()
	err := json.NewEncoder(bf).Encode(jobinf)
	//buf, err := json.Marshal(jobinf)
	if err != nil {
		util.FreeBuffer(bf)
		log.WithError(err).Error("序列化状态信息失败")
		return err
	}
	if _jobStatusProduce == nil {
		produce, err := rabbitmq.DefaultRMQ.RegisterProcedure(RobotJobStatusExchange)
		if err != nil {
			util.FreeBuffer(bf)
			log.WithError(err).Error("生产者创建失败")
			return err
		}
		_jobStatusProduce = produce
	}
	err = _jobStatusProduce.PublishSimple(string(RjrStatus), bf.Bytes())
	util.FreeBuffer(bf)
	return err
}

//SubscribeRobotJobStatus 订阅任务状态消息.
func SubscribeRobotJobStatus(queueName string, handler func(time time.Time, data ...RobotJobStatus)) error {
	if _jobStatusConsume == nil {
		consume, err := rabbitmq.DefaultRMQ.RegisterConsume(RobotJobStatusExchange, queueName, true, func(d amqp.Delivery) {
			jobStatuss := make([]RobotJobStatus, 0, 8)
			err := json.Unmarshal(d.Body, &jobStatuss)
			if err != nil {
				log.WithError(err).Error("接收机器人状态变更，数据解析失败")
				return
			}
			handler(time.Now(), jobStatuss...)
		})
		if err != nil {
			return err
		}
		_jobStatusConsume = consume
		consume.Subscribe(string(RjrStatus))
	}
	return nil
}
