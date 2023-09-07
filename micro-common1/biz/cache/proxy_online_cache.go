package cache

import (
	"common/redis"
	"context"
	"fmt"
	"strings"
	"time"
)

//前置机在线缓存
const (
	proxyOnlineKey     = "proxy_online:%s"
	proxyOnlineTimeOUt = time.Minute
)

//刷新前置机在线状态
func UpdateProxyOnline(officeId string) error {
	return redis.Set(context.TODO(), getProxyOnlineKey(officeId), time.Now().Unix(), proxyOnlineTimeOUt).Err()
}

//获取所有在线前置机的officeID
func FindProxyOnlineAll() ([]string, error) {
	cmd := redis.Keys(context.TODO(), getProxyOnlineKey("*"))
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	result := make([]string, len(cmd.Val()))
	for i, key := range cmd.Val() {
		result[i] = key[strings.LastIndex(key, ":")+1:]
	}
	return result, nil
}

func getProxyOnlineKey(officeID string) string {
	return fmt.Sprintf(proxyOnlineKey, officeID)
}
