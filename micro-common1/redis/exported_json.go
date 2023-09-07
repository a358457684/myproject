package redis

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"
)

func SetJson(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cmd := Client.Set(ctx, key, bytes, expiration)
	logError(cmd)
	return cmd.Err()
}

func GetJson(ctx context.Context, v interface{}, key string) error {
	cmd := Client.Get(ctx, key)
	logError(cmd)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	bytes, err := cmd.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

func MGetJson(ctx context.Context, data interface{}, keys ...string) error {
	cmd := Client.MGet(ctx, keys...)
	logError(cmd)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()
	if dataType.Kind() != reflect.Ptr {
		return errors.New("请使用指针接收数据")
	}
	dataValue = dataValue.Elem()
	dataType = dataType.Elem()
	if dataType.Kind() != reflect.Slice {
		return errors.New("请使用Slice进行接收")
	}
	itemType := dataType.Elem()
	slice := make([]reflect.Value, 0)
	for _, v := range cmd.Val() {
		item := reflect.New(itemType)
		_ = json.Unmarshal([]byte(v.(string)), item.Interface())
		slice = append(slice, item.Elem())
	}
	s := reflect.Append(dataValue, slice...)
	dataValue.Set(s)
	return nil
}

func HSetJson(ctx context.Context, key string, field string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cmd := Client.HSet(ctx, key, field, bytes)
	logError(cmd)
	return cmd.Err()
}

func HGetJson(ctx context.Context, v interface{}, key string, field string) error {
	cmd := Client.HGet(ctx, key, field)
	logError(cmd)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	bytes, err := cmd.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

func HGetALLJson(ctx context.Context, v interface{}, key string) error {
	cmd := Client.HGetAll(ctx, key)
	logError(cmd)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	vt := reflect.TypeOf(v).Elem()
	result := reflect.ValueOf(v)
	for k, v := range cmd.Val() {
		item := reflect.New(vt).Interface()
		_ = json.Unmarshal([]byte(v), item)
		result.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(item).Elem())
	}
	return nil
}

func LPushJson(ctx context.Context, key string, values ...interface{}) error {
	byteArray := make([]interface{}, len(values))
	for i, value := range values {
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		byteArray[i] = bytes
	}
	cmd := Client.LPush(ctx, key, byteArray...)
	logError(cmd)
	return cmd.Err()
}

func LRangeJson(ctx context.Context, v interface{}, key string, start, stop int64) error {
	cmd := Client.LRange(ctx, key, start, stop)
	logError(cmd)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	vt := reflect.TypeOf(v)
	if vt.Kind() != reflect.Ptr {
		return errors.New("请传入指针类型数组")
	}
	vt = vt.Elem().Elem()
	result := reflect.ValueOf(v).Elem()
	newResult := result
	for _, v := range cmd.Val() {
		item := reflect.New(vt).Interface()
		_ = json.Unmarshal([]byte(v), item)
		newResult = reflect.Append(newResult, reflect.ValueOf(item).Elem())
	}
	result.Set(newResult)
	return nil
}
