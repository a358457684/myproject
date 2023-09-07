package cache

import (
	"common/biz/enum"
	"common/redis"
	"context"
	"fmt"
	"time"
)

const (
	monitorScopeKey = "monitor_scope:%s:%s"
)

type MonitorScopeVo struct {
	X         float64
	Y         float64
	Status    enum.RobotStatusEnum
	PushCount int
	Time      time.Time
}

func SaveMonitorScope(officeId, robotId string, monitorScopeVo MonitorScopeVo) error {
	return redis.SetJson(context.TODO(), getMonitorScopeKey(officeId, robotId), monitorScopeVo, 0)
}

func DelMonitorScope(officeId, robotId string) error {
	return redis.Del(context.TODO(), getMonitorScopeKey(officeId, robotId)).Err()
}

func GetMonitorScope(officeId, robotId string) (data MonitorScopeVo, err error) {
	err = redis.GetJson(context.TODO(), &data, getMonitorScopeKey(officeId, robotId))
	return
}

func FindMonitorScopeAll() (officeRobotKeys []OfficeRobotKey, data []MonitorScopeVo, err error) {
	ctx := context.TODO()
	officeRobotKeys, err = findMonitorScopeKey(ctx, "*", "*")
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

func findMonitorScopeKey(ctx context.Context, officeId, robotId string) (keys []OfficeRobotKey, err error) {
	cmd := redis.Keys(ctx, getMonitorScopeKey(officeId, robotId))
	err = cmd.Err()
	for _, key := range cmd.Val() {
		keys = append(keys, OfficeRobotKey{Data: key})
	}
	return
}

func getMonitorScopeKey(officeId, robotId string) string {
	return fmt.Sprintf(monitorScopeKey, officeId, robotId)
}
