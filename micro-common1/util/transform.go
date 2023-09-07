package util

import (
	"encoding/json"
	"fmt"
	"reflect"
)

//结构体转json
func StructToJson(input interface{}) (map[string]interface{}, error) {
	msg, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(msg, &result)
	return result, err
}

//结构体转json，不返回error
func StructToJsonnoerr(input interface{}) map[string]interface{} {
	msg, err := json.Marshal(input)
	if err != nil {
		return nil
	}
	var result map[string]interface{}
	err = json.Unmarshal(msg, &result)
	return result
}

//json转结构体
func Jsontostruct(input map[string]interface{}, output interface{}) error {

	msg, err := json.Marshal(input)
	if err != nil {
		return err
	}

	err = json.Unmarshal(msg, &output)
	return err
}

//json interface转结构体
func Interfacetostruct(input interface{}, output interface{}) error {

	msg, err := json.Marshal(input)
	if err != nil {
		return err
	}

	err = json.Unmarshal(msg, &output)
	return err
}

//结构体同名字段拷贝
func CopyFields(dest interface{}, src interface{}, fields ...string) (err error) {
	at := reflect.TypeOf(dest)
	av := reflect.ValueOf(dest)
	bt := reflect.TypeOf(src)
	bv := reflect.ValueOf(src)

	// 简单判断下
	if at.Kind() != reflect.Ptr {
		err = fmt.Errorf("a must be a struct pointer")
		return
	}
	av = reflect.ValueOf(av.Interface())

	// 要复制哪些字段
	_fields := make([]string, 0)
	if len(fields) > 0 {
		_fields = fields
	} else {
		for i := 0; i < bv.NumField(); i++ {
			_fields = append(_fields, bt.Field(i).Name)
		}
	}

	if len(_fields) == 0 {
		return
	}

	// 复制
	for i := 0; i < len(_fields); i++ {
		name := _fields[i]
		f := av.Elem().FieldByName(name)
		bValue := bv.FieldByName(name)

		// a中有同名的字段并且类型一致才复制
		if f.IsValid() && f.Kind() == bValue.Kind() && f.Kind() != reflect.Struct {
			f.Set(bValue)
		}
	}
	return
}
