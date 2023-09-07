package util

import "reflect"

type StringInt struct {
	String string
	Int    int
}

func NewStringInt(String string, Int int) StringInt {
	return StringInt{
		String: String,
		Int:    Int,
	}
}

func NewStringIntSlice(mapData map[string]int) []StringInt {
	data := make([]StringInt, len(mapData))
	i := 0
	for k, v := range mapData {
		data[i] = NewStringInt(k, v)
		i += 1
	}
	return data
}

func CheckFieldExist(src interface{}, field string) bool {
	t := reflect.TypeOf(src)
	if _, ok := t.FieldByName(field); ok {
		return true
	} else {
		return false
	}
}
