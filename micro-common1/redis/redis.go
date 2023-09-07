package redis

import (
	"common/config"
	"common/log"
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

var Client redis.UniversalClient

func initRedis(options *config.RedisOptions) error {
	log.Info("开始初始化Redis...")
	var result string
	var err error
	var model string
	var ctx = context.Background()
	if options.Single != nil {
		model = "单例模式"
		Client = redis.NewClient(options.Single)
		result, err = Client.Ping(ctx).Result()
	} else if options.Sentinel != nil {
		model = "哨兵模式"
		Client = redis.NewFailoverClient(options.Sentinel)
		result, err = Client.Ping(ctx).Result()
	} else if options.Cluster != nil {
		model = "集群模式"
		Client = redis.NewClusterClient(options.Cluster)
		result, err = Client.Ping(ctx).Result()
	} else {
		err = errors.New("未知的Redis配置模式")
	}
	log.Infof("Redis连接模式：%s", model)
	if result != "PONG" {
		err = errors.New("Redis验证失败：" + result)
	}
	return err
}
