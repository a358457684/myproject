package cache

import (
	"common/biz/enum"
	"common/redis"
	"context"
	"fmt"
	"time"
)

const (
	monitorStatusKey = "monitor_status:%s:%s"
)

type MonitorStatusVo struct {
	Status    enum.RobotStatusEnum
	PushCount int
	Time      time.Time
}

func SaveMonitorStatus(officeId, robotId string, monitorStatusVo MonitorStatusVo) error {
	return redis.SetJson(context.TODO(), getMonitorStatusKey(officeId, robotId), monitorStatusVo, 0)
}

func DelMonitorStatus(officeId, robotId string) error {
	return redis.Del(context.TODO(), getMonitorStatusKey(officeId, robotId)).Err()
}

func GetMonitorStatus(officeId, robotId string) (data MonitorStatusVo, err error) {
	err = redis.GetJson(context.TODO(), &data, getMonitorStatusKey(officeId, robotId))
	return
}

func FindMonitorStatusAll() (officeRobotKeys []OfficeRobotKey, data []MonitorStatusVo, err error) {
	ctx := context.TODO()
	officeRobotKeys, err = findMonitorStatusKey(ctx, "*", "*")
	if err != nil {
		return
	}
	var keys []string
	for _, officeRobotKey := range officeRobotKeys {
		keys = append(keys, officeRobotKey.Data)
	}
	err = redis.MGetJson(ctx, &data, keys...)
	return
}

func findMonitorStatusKey(ctx context.Context, officeId, robotId string) (keys []OfficeRobotKey, err error) {
	cmd := redis.Keys(ctx, getMonitorStatusKey(officeId, robotId))
	err = cmd.Err()
	for _, key := range cmd.Val() {
		keys = append(keys, OfficeRobotKey{Data: key})
	}
	return
}

func getMonitorStatusKey(officeId, robotId string) string {
	return fmt.Sprintf(monitorStatusKey, officeId, robotId)
}
