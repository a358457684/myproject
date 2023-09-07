package util

import (
	"reflect"
)

type Stinfo struct {
	Field  string
	Value  interface{}
	Handle func(string) interface{}
}

func FilterOr(st interface{}, infos ...Stinfo) interface{} {
	var list []interface{}

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Slice {
		return nil
	}

	n := v.Len()
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		for _, info := range infos {
			avalue := v.Index(i).FieldByName(info.Field)
			if info.Handle == nil {
				if avalue.String() == info.Value {
					list = append(list, v.Index(i).Interface())
					continue
				}
			} else {
				if info.Handle(avalue.String()) == info.Value {
					list = append(list, v.Index(i).Interface())
					continue
				}
			}

		}
	}

	return list
}

func FilterAnd(st interface{}, infos ...Stinfo) interface{} {
	var list []interface{}

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Slice {
		return nil
	}

	n := v.Len()
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		for _, info := range infos {
			avalue := v.Index(i).FieldByName(info.Field)
			if info.Handle == nil {
				if avalue.String() != info.Value {
					continue
				}
			} else {
				if info.Handle(avalue.String()) != info.Value {
					continue
				}
			}
			list = append(list, v.Index(i).Interface())
		}
	}

	return list
}

func FilterAny(st interface{}, infos ...Stinfo) int {
	var list []interface{}

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Slice {
		return -1
	}

	n := v.Len()
	if n == 0 {
		return -1
	}

	for i := 0; i < n; i++ {
		for _, info := range infos {
			avalue := v.Index(i).FieldByName(info.Field)
			if avalue.String() == info.Value {
				list = append(list, v.Index(i).Interface())
				return i
			}
		}
	}

	return -1
}

func FilterfuncAny(st interface{}, infos ...Stinfo) int {

	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Slice {
		return -1
	}

	n := v.Len()
	if n == 0 {
		return -1
	}

	for i := 0; i < n; i++ {
		for _, info := range infos {
			avalue := v.Index(i).FieldByName(info.Field)
			if info.Handle == nil {
				if avalue.String() != info.Value {
					continue
				}
			} else {
				if info.Handle(avalue.String()) != info.Value {
					continue
				}
			}
			return i
		}
	}

	return -1
}

func FilterCount(st interface{}, infos ...Stinfo) int {
	count := 0
	v := reflect.ValueOf(st)
	if v.Kind() != reflect.Slice {
		return count
	}

	n := v.Len()
	if n == 0 {
		return count
	}

	for i := 0; i < n; i++ {
		for _, info := range infos {
			avalue := v.Index(i).FieldByName(info.Field)
			if info.Handle != nil {
				if info.Handle(avalue.String()) == info.Value {
					count++
				}
			}
		}
	}

	return count
}
