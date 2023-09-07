package cache

import (
	"common/biz/enum"
	"common/redis"
	"context"
	"fmt"
	"time"
)

const (
	monitorJobKey = "monitor_job:%s:%s"
)

type MonitorJobVo struct {
	JobId     string
	Status    enum.RobotStatusEnum
	PushCount int
	Time      time.Time
}

func SaveMonitorJob(officeId, robotId string, monitorJobVo MonitorJobVo) error {
	return redis.SetJson(context.TODO(), getMonitorJobKey(officeId, robotId), monitorJobVo, 0)
}

func DelMonitorJob(officeId, robotId string) error {
	return redis.Del(context.TODO(), getMonitorJobKey(officeId, robotId)).Err()
}

func GetMonitorJob(officeId, robotId string) (data MonitorJobVo, err error) {
	err = redis.GetJson(context.TODO(), &data, getMonitorJobKey(officeId, robotId))
	return
}

func FindMonitorJobAll() (officeRobotKeys []OfficeRobotKey, data []MonitorJobVo, err error) {
	ctx := context.TODO()
	officeRobotKeys, err = findMonitorJobKey(ctx, "*", "*")
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

func findMonitorJobKey(ctx context.Context, officeId, robotId string) (keys []OfficeRobotKey, err error) {
	cmd := redis.Keys(ctx, getMonitorJobKey(officeId, robotId))
	err = cmd.Err()
	for _, key := range cmd.Val() {
		keys = append(keys, OfficeRobotKey{Data: key})
	}
	return
}

func getMonitorJobKey(officeId, robotId string) string {
	return fmt.Sprintf(monitorJobKey, officeId, robotId)
}
