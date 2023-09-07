package cache

import (
	"common/log"
	"common/redis"
	"common/util"
	"context"
	"time"
)

const (
	WorkOrderKey     = "admin:workOrder:"
	WorkOrderMessage = "admin:workOrderMessage:"
	timeOut          = time.Hour * 24 * 30
)

type WorkOrderMessageDTO struct {
	ID      string    `json:"id"`      //工单ID
	Message string    `json:"message"` //工单消息内容
	Time    time.Time `json:"time"`    //时间
}

func AddWorkOrder(userID, workOrderID string) {
	if err := redis.SAdd(context.Background(), WorkOrderKey+userID, workOrderID).Err(); err != nil {
		log.WithError(err).Error("添加用户工单缓存失败")
	}
	redis.Expire(context.Background(), WorkOrderKey+userID, timeOut)
}

func DelWorkOrder(userID, workOrderID string) {
	if err := redis.SRem(context.Background(), WorkOrderKey+userID, workOrderID).Err(); err != nil {
		log.WithError(err).Error("删除用户工单缓存失败")
	}
	redis.Expire(context.Background(), WorkOrderKey+userID, timeOut)
}

func FindWorkOrderMessage(userID string) ([]WorkOrderMessageDTO, error) {
	workOrderIDs, err := redis.SMembers(context.Background(), WorkOrderKey+userID).Result()
	if err != nil {
		return nil, util.WrapErr(err, "获取用户工单列表失败")
	}
	size := len(workOrderIDs)
	if size == 0 {
		return []WorkOrderMessageDTO{}, nil
	}
	keys := make([]string, size)
	for i, workOrderID := range workOrderIDs {
		keys[i] = WorkOrderMessage + workOrderID
	}
	var result []WorkOrderMessageDTO
	if err := redis.MGetJson(context.Background(), &result, keys...); err != nil {
		return nil, util.WrapErr(err, "查询工单消息失败")
	}
	redis.Expire(context.Background(), WorkOrderKey+userID, timeOut)
	return result, nil
}

func AddWorkOrderMessage(workOrderID, message string) {
	messageVO := WorkOrderMessageDTO{
		ID:      workOrderID,
		Message: message,
		Time:    time.Now(),
	}
	if err := redis.SetJson(context.Background(), WorkOrderMessage+workOrderID, messageVO, timeOut); err != nil {
		log.WithError(err).Error("添加工单消息失败")
	}
}
