package cache

import (
	"common/biz/dto"
	"common/redis"
	"context"
	"fmt"
)

const (
	robotStatusKey = "robot_status:%s:%s"
)

func HasRobotStatus(officeId, robotId string) (bool, error) {
	if officeId == "*" {
		cmd := redis.Keys(context.TODO(), getRobotStatusKey(officeId, robotId))
		return len(cmd.Val()) > 0, cmd.Err()
	}
	cmd := redis.Exists(context.TODO(), getRobotStatusKey(officeId, robotId))
	return cmd.Val() > 0, cmd.Err()
}

func SaveRobotStatus(officeId, robotId string, vo dto.RobotStatus) error {
	return redis.SetJson(context.TODO(), getRobotStatusKey(officeId, robotId), vo, 0)
}

func GetRobotStatus(officeId, robotId string) (result dto.RobotStatus, err error) {
	err = redis.GetJson(context.TODO(), &result, getRobotStatusKey(officeId, robotId))
	return
}

func RemoveRobotStatus(officeId, robotId string) (int64, error) {
	count, err := redis.Del(context.Background(), getRobotStatusKey(officeId, robotId)).Result()
	return count, err
}

func FindRobotStatusByOfficeId(officeId string) (result []dto.RobotStatus, err error) {
	return finRobotStatus(officeId, "*")
}

func FindRobotStatusAll() (result []dto.RobotStatus, err error) {
	return finRobotStatus("*", "*")
}

func finRobotStatus(officeId, robotId string) (result []dto.RobotStatus, err error) {
	ctx := context.TODO()
	cmd := redis.Keys(ctx, getRobotStatusKey(officeId, robotId))
	if cmd.Err() != nil {
		err = cmd.Err()
		return
	}
	err = redis.MGetJson(ctx, &result, cmd.Val()...)
	return
}

//func FindRobotStatusOnline() (result []dto.RobotStatus, err error) {
//	officeRobotKeys, err1 := FindMonitorOnlineAll()
//	if err1 != nil {
//		err = err1
//		return
//	}
//	keys := make([]string, 0)
//	for _, officeRobotKey := range officeRobotKeys {
//		keys = append(keys, getRobotStatusKey(officeRobotKey.OfficeId(), officeRobotKey.RobotId()))
//	}
//	err = redis.MGetJson(context.TODO(), &result, keys...)
//	return
//}
//
//func FindRobotStatusOnlineByOfficeId(officeId string) (result []dto.RobotStatus, err error) {
//	officeRobotKeys, err1 := FindMonitorOnlineByOfficeId(officeId)
//	if err1 != nil {
//		err = err1
//		return
//	}
//	keys := make([]string, 0)
//	for _, officeRobotKey := range officeRobotKeys {
//		keys = append(keys, getRobotStatusKey(officeRobotKey.OfficeId(), officeRobotKey.RobotId()))
//	}
//	if len(keys) > 0 {
//		err = redis.MGetJson(context.TODO(), &result, keys...)
//	}
//	return
//}

func getRobotStatusKey(officeId, robotId string) string {
	return fmt.Sprintf(robotStatusKey, officeId, robotId)
}
