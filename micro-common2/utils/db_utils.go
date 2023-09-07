package utils

import (
	"errors"
	"reflect"
	"strings"
	"xorm.io/xorm"
)

// 分页插入数据库
func Insert(db *xorm.Engine, beans ...interface{}) error {
	return InsertWithTableNameAndSize(db, "", 500, beans...)
}

func InsertWithTableName(db *xorm.Engine, tableName string, beans ...interface{}) error {
	return InsertWithTableNameAndSize(db, tableName, 500, beans...)
}

func InsertWithSize(db *xorm.Engine, size int, beans ...interface{}) error {
	return InsertWithTableNameAndSize(db, "", size, beans...)
}

func InsertWithTableNameAndSize(db *xorm.Engine, tableName string, size int, beans ...interface{}) error {
	if nil == db {
		return errors.New(`db can not be nil`)
	}

	if nil == beans || len(beans) == 0 {
		return nil
	}

	trans := db.NewSession()
	defer trans.Close()
	_ = trans.Begin()

	err := TransInsertWithTableNameAndSize(trans, tableName, size, beans...)
	if nil != err {
		_ = trans.Rollback()
		return err
	}

	return trans.Commit()
}

func TransInsert(trans *xorm.Session, beans ...interface{}) error {
	return TransInsertWithTableNameAndSize(trans, "", 500, beans...)
}

func TransInsertWithTableName(trans *xorm.Session, tableName string, beans ...interface{}) error {
	return TransInsertWithTableNameAndSize(trans, tableName, 500, beans...)
}

func TransInsertWithSize(trans *xorm.Session, size int, beans ...interface{}) error {
	return TransInsertWithTableNameAndSize(trans, "", size, beans...)
}

func TransInsertWithTableNameAndSize(trans *xorm.Session, tableName string, size int, beans ...interface{}) error {
	if nil == trans {
		return errors.New(`trans can not be nil`)
	}

	if nil == beans || len(beans) == 0 {
		return nil
	}
	tableName = strings.TrimSpace(tableName)

	if size <= 0 {
		size = 500
	}

	var objects []interface{}
	for _, bean := range beans {
		sliceValue := reflect.Indirect(reflect.ValueOf(bean))
		if sliceValue.Kind() == reflect.Slice {
			sLen := sliceValue.Len()
			if sLen == 0 {
				continue
			}
			if sLen <= size {
				objects = append(objects, bean)
				continue
			}
			idx := 0
			arrMap := make(map[int][]interface{})
			for i := 0; i < sliceValue.Len(); i++ {
				if i%size == 0 {
					idx++
				}
				v := sliceValue.Index(i)
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}
				elemValue := v.Interface()
				if nil == arrMap[idx] {
					arrMap[idx] = []interface{}{elemValue}
				} else {
					arrMap[idx] = append(arrMap[idx], elemValue)
				}
			}
			for i := 1; i <= idx; i++ {
				arr := arrMap[i]
				objects = append(objects, &arr)
			}
		} else {
			objects = append(objects, bean)
		}
	}

	var err error
	if tableName != "" {
		for i := 0; i < len(objects); i++ {
			_, err = trans.Table(tableName).Insert(objects[i])
			if nil != err {
				return err
			}
		}
	} else {
		_, err = trans.Insert(objects...)
		if nil != err {
			return err
		}
	}
	return nil
}
