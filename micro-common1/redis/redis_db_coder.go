package redis

import (
	"github.com/suiyunonghen/DxCommonLib"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RedisTable interface {
	TableName()string
}

type RedisValueCoder interface {
	Encode()string
	EncodeHash()(string)
	Decode(key, redisValue string)error
}

var(
	RedisTableInter = reflect.TypeOf((*RedisTable)(nil)).Elem()
	RedisValueCoderType = reflect.TypeOf((*RedisValueCoder)(nil)).Elem()
	TimeType = reflect.TypeOf((*time.Time)(nil)).Elem()
)

var(
	structFields		sync.Map			//map[reflect.Type]*redisFields
)

type redisField struct {
	isKey				bool
	index				uint16
	fieldName			string
	fieldType			reflect.Type
}

//是否是内嵌结构
func (field *redisField)IsInnerStruct()bool  {
	return field.fieldType.Name() == field.fieldName
}

type redisKeyIndex struct {
	fieldIndex		uint16
	RealFieldIndex	uint16
}

type redisFields struct{
	redisKeyIndex	[]redisKeyIndex
	fields 			[]redisField
}

func (fields *redisFields)Len() int  {
	return len(fields.redisKeyIndex)
}

func (fields *redisFields)Less(i, j int) bool  {
	return fields.redisKeyIndex[i].fieldIndex < fields.redisKeyIndex[j].fieldIndex
}

func (fields *redisFields)Swap(i, j int)  {
	fields.redisKeyIndex[i],fields.redisKeyIndex[j] = fields.redisKeyIndex[j],fields.redisKeyIndex[i]
}

func (fields *redisFields)IfKeyIndexExists(fieldName string)int  {
	for i := 0;i<len(fields.redisKeyIndex);i++{
		if fields.fields[int(fields.redisKeyIndex[i].fieldIndex)].fieldName == fieldName{
			return i
		}
	}
	return -1
}

func (fields *redisFields)CreateIndexKey(queryMap map[string]string)string  {
	keys := make([]string,0,3)
	for i := 0;i<len(fields.redisKeyIndex);i++{
		queryValue,ok := queryMap[fields.fields[fields.redisKeyIndex[i].fieldIndex].fieldName]
		if ok{
			keys = append(keys,queryValue)
		}else{
			keys = append(keys,"*")
		}
	}
	return strings.Join(keys,":")
}

func (fields *redisFields)hashKeymap(v reflect.Value,mp map[string]string)  {
	for i := 0;i<len(fields.redisKeyIndex);i++ {
		f := v.Field(int(fields.redisKeyIndex[i].RealFieldIndex))
		switch f.Kind() {
		case reflect.String:
			str := f.String()
			if str != ""{
				mp[fields.fields[fields.redisKeyIndex[i].fieldIndex].fieldName] = str
			}
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			mp[fields.fields[fields.redisKeyIndex[i].fieldIndex].fieldName] = strconv.Itoa(int(f.Int()))
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			mp[fields.fields[fields.redisKeyIndex[i].fieldIndex].fieldName] = strconv.Itoa(int(f.Uint()))
		case reflect.Bool:
			if f.Bool(){
				mp[fields.fields[fields.redisKeyIndex[i].fieldIndex].fieldName] = "true"
			}else{
				mp[fields.fields[fields.redisKeyIndex[i].fieldIndex].fieldName] = "false"
			}
		case reflect.Struct:
			if fields.fields[int(fields.redisKeyIndex[i].fieldIndex)].IsInnerStruct(){ //内嵌结构体
				fldtype := fieldsByType(fields.fields[int(fields.redisKeyIndex[i].fieldIndex)].fieldType)
				fldtype.hashKeymap(f,mp)
			}
		}
	}
}

func (fields *redisFields)hashKey(v reflect.Value)string  {
	var builder strings.Builder
	builder.Grow(128)
	isfirst := true
	for i := 0;i<len(fields.redisKeyIndex);i++ {
		if !isfirst{
			builder.WriteByte(':')
		}else{
			isfirst = false
		}
		f := v.Field(int(fields.redisKeyIndex[i].RealFieldIndex))
		switch f.Kind() {
		case reflect.String:
			builder.WriteString(f.String())
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			builder.WriteString(strconv.Itoa(int(f.Int())))
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			builder.WriteString(strconv.Itoa(int(f.Uint())))
		case reflect.Bool:
			if f.Bool(){
				builder.WriteString("true")
			}else{
				builder.WriteString("false")
			}
		case reflect.Struct:
			if fields.fields[int(fields.redisKeyIndex[i].fieldIndex)].IsInnerStruct(){ //内嵌结构体
				fld := fieldsByType(fields.fields[int(fields.redisKeyIndex[i].fieldIndex)].fieldType)
				builder.WriteString(fld.hashKey(f))
			}
		}
	}
	return builder.String()
}

//redisDB: "fieldName,KeyIndex"
/*
type RedisTestStruct struct {
	Name		string		`redisDB:"Name,1"`
	ID			string 		`redisDB:"ID,0"`
    Age			int
}

func (RedisTestStruct) TableName()string{
  return "RedisTest"
}

最终将构建一个 RedisTest作为Key的HashMap的数据结构到Redis
Key值将为  ID:Name	RedisTestStruct的去掉索引键值的Json字符串
RedisTestStruct{"不得闲","234",23}
最终在Redis中为
Key=234:不得闲， Value={"Age":23}

然后就可以支持根据键值查询，比如查询名称为 TT的就是
Key=*:TT,模糊查询
*/
func fieldsByType(typ reflect.Type)*redisFields  {
	if v, ok := structFields.Load(typ); ok {
		return v.(*redisFields)
	}
	fieldCount := typ.NumField()
	var result redisFields
	var field redisField
	result.fields = make([]redisField,0,fieldCount)
	realFieldIndex := make(map[uint16]uint16,fieldCount)
	for i := 0;i<fieldCount;i++{
		f := typ.Field(i)
		namebt := DxCommonLib.FastString2Byte(f.Name)
		if namebt[0] < 'A' || namebt[0] > 'Z'{ //去掉小写
			continue
		}
		tagStr := f.Tag.Get("redisDB")
		field.index = uint16(i)
		field.isKey = false
		field.fieldType = f.Type
		if tagStr != ""{
			fields := strings.Split(tagStr,",")
			if len(fields) == 2 && len(fields[1]) > 0{
				keyIndex := int(DxCommonLib.StrToIntDef(fields[1],-1))
				if keyIndex > -1{
					//newint := longInt{uint16(keyIndex),uint16(i)}
					result.redisKeyIndex = append(result.redisKeyIndex,redisKeyIndex{uint16(keyIndex),uint16(i)})
					field.isKey = true
					realFieldIndex[uint16(keyIndex)] = uint16(len(result.fields))
				}
			}
			if fields[0] == ""{
				field.fieldName = f.Name
			}else{
				field.fieldName = fields[0]
			}
			result.fields = append(result.fields,field)
		}else{
			if f.Type.Kind() == reflect.Struct{
				if f.Type.Name() != "RWMutex"{
					field.fieldName = f.Name
					result.fields = append(result.fields,field)
					//注册这个类型
					fieldsByType(f.Type)
				}
			}else{
				field.fieldName = f.Name
				result.fields = append(result.fields,field)
			}
		}
	}
	//将rediskeyindex根据keyIndex排序
	sort.Sort(&result)
	for i := 0;i<len(result.redisKeyIndex);i++{
		result.redisKeyIndex[i].fieldIndex = realFieldIndex[result.redisKeyIndex[i].fieldIndex] //当前结构的位置索引
	}
	structFields.Store(typ,&result)
	return &result
}

