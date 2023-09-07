package util

import (
	"reflect"
	"sort"
)

//string slice 逆转
func SliceReverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func Contains(src []string, sub []string) bool {

	if len(src) < len(sub) {
		return false
	}
	smap := make(map[string]int)
	for _, item := range src {
		smap[item] = 1
	}
	for _, item := range sub {
		if smap[item] == 0 {
			return false
		}
	}
	return true
}

func ContainOne(src []string, sub string) bool {

	for _, item := range src {
		if item == sub {
			return true
		}
	}
	return false
}

func Dump(src []string) []string {
	result := []string{}
	mp := make(map[string]bool)
	for _, item := range src {
		if mp[item] == false {
			result = append(result, item)
			mp[item] = true
		}
	}
	return result
}

func Removesub(src []string, sub []string) []string {
	var result = []string{}
	mp := make(map[string]bool)
	for _, item := range sub {
		mp[item] = true
	}
	for _, item := range src {
		if mp[item] == false {
			result = append(result, item)
		}
	}
	return result
}

//通用排序
//结构体排序，必须重写数组Len() Swap() Less()函数
type body_wrapper struct {
	Bodys []interface{}
	by    func(p, q *interface{}) bool //内部Less()函数会用到
}
type SortBodyBy func(p, q *interface{}) bool //定义一个函数类型

//数组长度Len()
func (acw body_wrapper) Len() int {
	return len(acw.Bodys)
}

//元素交换
func (acw body_wrapper) Swap(i, j int) {
	acw.Bodys[i], acw.Bodys[j] = acw.Bodys[j], acw.Bodys[i]
}

//比较函数，使用外部传入的by比较函数
func (acw body_wrapper) Less(i, j int) bool {
	return acw.by(&acw.Bodys[i], &acw.Bodys[j])
}

//自定义排序字段，参考SortBodyByCreateTime中的传入函数
func SortBody(bodys []interface{}, by SortBodyBy) {
	sort.Sort(body_wrapper{bodys, by})
}

//按照createtime排序，需要注意是否有createtime
func SortBodyByCreateTime(bodys []interface{}) {
	sort.Sort(body_wrapper{bodys, func(p, q *interface{}) bool {
		v := reflect.ValueOf(*p)
		i := v.FieldByName("Create_time")
		v = reflect.ValueOf(*q)
		j := v.FieldByName("Create_time")
		return i.String() > j.String()
	}})
}
