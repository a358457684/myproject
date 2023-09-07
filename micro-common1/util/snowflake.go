package util

//增加一个生成雪花ID的功能模块
import (
	"common/config"
	"common/log"
	"common/redis"
	"context"
	"github.com/bwmarrin/snowflake"
	"github.com/suiyunonghen/DxCommonLib"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	node        *snowflake.Node
	atomicIndex int64
)

func init() {
	if config.Data.Project != nil {
		if config.Data.Project.WorkerId > 1023 {
			//采用redis获取配置信息
			if config.Data.Project.Name != "" {
				lockid := config.Data.Project.Name + ":lck"
				redis.SimpleLock(lockid)
				id, err := redis.Get(context.Background(), config.Data.Project.Name).Result()
				if err != nil {
					if redis.IsRedisNil(err) {
						config.Data.Project.WorkerId = 0
						redis.Set(context.Background(), config.Data.Project.Name, strconv.Itoa(int(config.Data.Project.WorkerId+1)), 0)
					}
				} else {
					lastid := DxCommonLib.StrToIntDef(id, -1)
					if lastid >= 0 {
						config.Data.Project.WorkerId = uint16(lastid)
						lastid++
						if lastid > 1023 {
							lastid = 0
						}
						redis.Set(context.Background(), config.Data.Project.Name, strconv.Itoa(int(lastid)), 0)
					}
				}
				redis.SimpleUnLock(lockid)
			}
			if config.Data.Project.WorkerId > 1023 {
				_, ip32 := GetWanNetIP()
				if ip32 > 0 {
					config.Data.Project.WorkerId = uint16(ip32 % 1024)
				}
			}
		}
		if config.Data.Project.WorkerId < 1024 {
			Node, err := snowflake.NewNode(int64(config.Data.Project.WorkerId))
			if err != nil {
				log.WithError(err).Debug("初始化雪花分布式ID生成失败，采用本地当前时间作为初始ID")
				atomicIndex = int64(time.Now().Nanosecond())
				return
			}
			node = Node
		}
	}
}

func GetSnowflakeID() int64 {
	if node != nil {
		return int64(node.Generate())
	} else {
		return atomic.AddInt64(&atomicIndex, 1)
	}
}
