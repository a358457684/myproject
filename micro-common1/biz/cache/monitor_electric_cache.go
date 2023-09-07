package cache

import (
	"common/redis"
	"context"
	"fmt"
	"time"
)

const (
	monitorElectricKey = "monitor_electric:%s:%s"
)

type MonitorElectricVo struct {
	Electric  float64
	PushCount int
	Time      time.Time
}

func SaveMonitorElectric(officeId, robotId string, monitorElectricVo MonitorElectricVo) error {
	return redis.SetJson(context.TODO(), getMonitorElectricKey(officeId, robotId), monitorElectricVo, 0)
}

func DelMonitorElectric(officeId, robotId string) error {
	return redis.Del(context.TODO(), getMonitorElectricKey(officeId, robotId)).Err()
}

func GetMonitorElectric(officeId, robotId string) (data MonitorElectricVo, err error) {
	err = redis.GetJson(context.TODO(), &data, getMonitorElectricKey(officeId, robotId))
	return
}

func FindMonitorElectricAll() (officeRobotKeys []OfficeRobotKey, data []MonitorElectricVo, err error) {
	ctx := context.TODO()
	officeRobotKeys, err = findMonitorElectricKey(ctx, "*", "*")
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

func findMonitorElectricKey(ctx context.Context, officeId, robotId string) (keys []OfficeRobotKey, err error) {
	cmd := redis.Keys(ctx, getMonitorElectricKey(officeId, robotId))
	err = cmd.Err()
	for _, key := range cmd.Val() {
		keys = append(keys, OfficeRobotKey{Data: key})
	}
	return
}

func getMonitorElectricKey(officeId, robotId string) string {
	return fmt.Sprintf(monitorElectricKey, officeId, robotId)
}
