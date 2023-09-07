package cache

import (
	"common/log"
	"common/redis"
	"common/util"
	"context"
	"fmt"
)

//对方法返回值进行缓存
const (
	functionDataKey  = "function:data:%s"
	functionTableKey = "function:table:%s"
)

// SaveFunctionCache 对函数执行结果进行缓存.
// functionName: 函数名称，全局唯一.
// tableKeys: 关键字,可以根据其中任意一个关键字删除其关联的方法缓存.
// data: 要缓存的数据.
// params: 调用函数时的参数.
func SaveFunctionCache(functionName string, tableKeys []string, data interface{}, params ...interface{}) error {
	dataKey := fmt.Sprintf(functionDataKey, functionName)
	if err := redis.HSetJson(context.TODO(), dataKey, fmt.Sprintf("%v", params), data); err != nil {
		return err
	}
	var result error
	for _, tableKey := range tableKeys {
		if err := redis.SAdd(context.TODO(), fmt.Sprintf(functionTableKey, tableKey), dataKey).Err(); err != nil {
			result = err
			log.WithError(err).Error("保存函数关键字数据失败")
		}
	}
	return result
}

// GetFunctionCache 查询函数缓存.
// result: 结果.
// functionName: 函数名称.
// params: 调用函数时的参数.
func GetFunctionCache(result interface{}, functionName string, params ...interface{}) error {
	return redis.HGetJson(context.TODO(), result, fmt.Sprintf(functionDataKey, functionName), fmt.Sprintf("%v", params))
}

// DelFunctionCache 根据关键字删除关联的函数缓存.
func DelFunctionCache(tableKeys ...string) error {
	var delKeys []string
	for _, tableKey := range tableKeys {
		tableKey = fmt.Sprintf(functionTableKey, tableKey)
		dataKeys, err := redis.SMembers(context.TODO(), tableKey).Result()
		if err != nil && !redis.IsRedisNil(err) {
			return util.WrapErr(err, "查询函数缓存数据失败")
		}
		delKeys = append(delKeys, dataKeys...)
		delKeys = append(delKeys, tableKey)
	}
	if err := redis.Del(context.TODO(), delKeys...).Err(); err != nil {
		return util.WrapErr(err, "删除函数缓存数据失败")
	}
	return nil
}

// DelAllFunctionCache 删除所有函数缓存数据.
func DelAllFunctionCache() error {
	dataKeys, err := redis.Keys(context.TODO(), fmt.Sprintf(functionDataKey, "*")).Result()
	if err != nil && !redis.IsRedisNil(err) {
		return util.WrapErr(err, "查询函数缓存数据条目失败")
	}
	tableKeys, err := redis.Keys(context.TODO(), fmt.Sprintf(functionTableKey, "*")).Result()
	if err != nil && !redis.IsRedisNil(err) {
		return util.WrapErr(err, "查询函数缓存关键字条目失败")
	}
	delKeys := append(dataKeys, tableKeys...)
	if err := redis.Del(context.TODO(), delKeys...).Err(); err != nil {
		return util.WrapErr(err, "删除函数缓存数据失败")
	}
	return nil
}
