package utils

import (
	"context"
	"micro-common1/log"
	"micro-common1/redis"
	"time"
)

// 普通业务使用
func GetLock(key string, timeout time.Duration) bool {
	ctx := context.Background()
	res, err := redis.SetNX(ctx, key, 1, timeout).Result()
	if err != nil {
		log.WithError(err).Errorln("====== 获取分布式锁异常 ======")
	}
	return res
}
