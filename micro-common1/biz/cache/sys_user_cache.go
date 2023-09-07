package cache

import (
	"common/log"
	"common/redis"
	"context"
)

const (
	sysUserKey = "admin:user"
)

func GetSysUser(user interface{}, userID string) error {
	return redis.HGetJson(context.TODO(), user, sysUserKey, userID)
}

func SaveSysUser(userID string, user interface{}) error {
	return redis.HSetJson(context.TODO(), sysUserKey, userID, user)
}

func DelSysUser(userID ...string) {
	if len(userID) == 0 {
		return
	}
	if err := redis.HDel(context.TODO(), sysUserKey, userID...).Err(); err != nil {
		log.WithError(err).Error("清除用户缓存失败")
	}
}

func ClearSysUser() {
	if err := redis.Del(context.TODO(), sysUserKey).Err(); err != nil {
		log.WithError(err).Error("清除用户缓存失败")
	}
}
