package cache

import (
	"common/redis"
	"context"
)

const (
	menuKey = "admin:menu"
)

func FindMenu(codes []string) ([]interface{}, error) {
	return redis.HMGet(context.TODO(), menuKey, codes...).Result()
}

func SaveMenu(code, name string) error {
	return redis.HSet(context.TODO(), menuKey, code, name).Err()
}

func ResetMenu(values []interface{}) error {
	if err := redis.Del(context.TODO(), menuKey).Err(); err != nil {
		return err
	}
	return redis.HMSet(context.TODO(), menuKey, values...).Err()
}
