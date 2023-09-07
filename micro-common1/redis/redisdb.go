package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/suiyunonghen/DxCommonLib"
	"github.com/suiyunonghen/dxsvalue"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func growSliceValue(v reflect.Value, n int) reflect.Value {
	diff := n - v.Len()
	if diff > 256 {
		diff = 256
	}
	v = reflect.AppendSlice(v, reflect.MakeSlice(v.Type(), diff, diff))
	return v
}

func decodeRedisObjValue(fvalue reflect.Value,value *dxsvalue.DxValue,tp reflect.Type,keys map[string]string)error  {
	if tp == nil{
		tp = fvalue.Type()
	}
	valueFields := fieldsByType(tp)
	for i := 0;i < len(valueFields.fields);i++{
		if valueFields.fields[i].isKey{
			if valueFields.fields[i].IsInnerStruct(){
				f := fvalue.Field(int(valueFields.fields[i].index))
				keyName := valueFields.fields[i].fieldName
				fvalue := value.ValueByName(keyName)
				err := decodeRedisObjValue(f,fvalue,valueFields.fields[i].fieldType,keys)
				if err != nil{
					return err
				}
			}else{
				decodeKeys(keys,&fvalue,valueFields,i)
			}
			continue
		}
		f := fvalue.Field(int(valueFields.fields[i].index))
		fvalue := value.ValueByName(valueFields.fields[i].fieldName)
		if fvalue == nil{
			continue
		}
		switch f.Kind() {
		case reflect.String:
			f.SetString(fvalue.String())
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			f.SetInt(fvalue.Int())
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			f.SetUint(uint64(fvalue.Int()))
		case reflect.Bool:
			f.SetBool(fvalue.Bool())
		case reflect.Float64,reflect.Float32:
			f.SetFloat(fvalue.Double())
		case reflect.Struct:
			if valueFields.fields[i].fieldType == TimeType{
				f.Set(reflect.ValueOf(fvalue.GoTime()))
			}else{
				err := decodeRedisObjValue(f,fvalue,valueFields.fields[i].fieldType,keys)
				if err != nil{
					return err
				}
			}
		}
	}
	return nil
}

func decodeKeys(keys map[string]string,v *reflect.Value,valuetype *redisFields,keyindex int) {
	f := v.Field(int(valuetype.fields[keyindex].index))
	kvalue,ok := keys[valuetype.fields[keyindex].fieldName]
	if !ok{
		kvalue = ""
	}
	switch f.Kind() {
	case reflect.String:
		f.SetString(kvalue)
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
		f.SetInt(DxCommonLib.StrToIntDef(kvalue,0))
	case reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		f.SetUint(uint64(DxCommonLib.StrToIntDef(kvalue,0)))
	case reflect.Float32, reflect.Float64:
		f.SetFloat(DxCommonLib.StrToFloatDef(kvalue,0))
	case reflect.Bool:
		if strings.ToUpper(kvalue) == "TRUE"{
			f.SetBool(true)
		}else{
			f.SetBool(false)
		}
	}
}

func decodeKeys2mapkeys(keys []string,mapkeys map[string]string,valuetype *redisFields)  {
	klen := len(keys)
	for i := 0;i<len(valuetype.redisKeyIndex);i++{
		if valuetype.fields[valuetype.redisKeyIndex[i].fieldIndex].IsInnerStruct(){
			newvtype := fieldsByType(valuetype.fields[valuetype.redisKeyIndex[i].fieldIndex].fieldType)
			decodeKeys2mapkeys(keys[i:],mapkeys,newvtype)
		}else if i < klen{
			mapkeys[valuetype.fields[valuetype.redisKeyIndex[i].fieldIndex].fieldName] = keys[i]
		}else{
			mapkeys[valuetype.fields[valuetype.redisKeyIndex[i].fieldIndex].fieldName] = ""
		}
	}
}

func decodeRedis(v *reflect.Value,key,redis string,valuetype *redisFields)error  {
	tp := v.Type()
	if valuetype == nil && tp.Kind() != reflect.Map{
		return errors.New("无法decode到目标")
	}
	if valuetype != nil && len(valuetype.fields) == 0{
		return nil
	}

	value,err := dxsvalue.NewValueFromJson(DxCommonLib.FastString2Byte(redis),true,true)
	if err != nil{
		return err
	}
	defer dxsvalue.FreeValue(value)
	keys := strings.Split(key,":")

	mapkeys := make(map[string]string,len(keys))
	decodeKeys2mapkeys(keys,mapkeys,valuetype)

	if valuetype != nil{ //结构体
		for i := 0;i<len(valuetype.fields);i++{
			if valuetype.fields[i].isKey{ //不解析key
				if valuetype.fields[i].IsInnerStruct(){
					f := v.Field(int(valuetype.fields[i].index))
					keyName := valuetype.fields[i].fieldName
					fvalue := value.ValueByName(keyName)
					err = decodeRedisObjValue(f,fvalue,valuetype.fields[i].fieldType,mapkeys)
					if err != nil{
						return err
					}
				}else{
					decodeKeys(mapkeys,v,valuetype,i)
				}
				continue
			}
			keyName := valuetype.fields[i].fieldName
			fvalue := value.ValueByName(keyName)
			if fvalue == nil{
				continue
			}
			f := v.Field(int(valuetype.fields[i].index))
			switch f.Kind(){
			case reflect.String:
				f.SetString(fvalue.String())
			case reflect.Int,reflect.Int64,reflect.Int8,reflect.Int16,reflect.Int32:
				f.SetInt(fvalue.Int())
			case reflect.Uint,reflect.Uint64,reflect.Uint8,reflect.Uint16,reflect.Uint32:
				f.SetUint(uint64(fvalue.Int()))
			case reflect.Float32,reflect.Float64:
				f.SetFloat(fvalue.Double())
			case reflect.Bool:
				f.SetBool(fvalue.Bool())
			case reflect.Slice:
				//返回的是否是[]byte数组
			case reflect.Struct:
				if valuetype.fields[i].fieldType == TimeType{
					f.Set(reflect.ValueOf(fvalue.GoTime()))
				}else if valuetype.fields[i].IsInnerStruct(){
					//内嵌结构内容解析
					err = decodeRedisObjValue(f,fvalue,valuetype.fields[i].fieldType,mapkeys)
					if err != nil{
						return err
					}
				}
			}
		}
		return nil
	}

	if tp.Kind() == reflect.Map{
		if tp.Key().Kind() != reflect.String{
			return errors.New("Key只能是string类型，无法decode到目标")
		}
		for i := 0;i<value.Count();i++{
			k,recordvalue := value.Items(i)
			if recordvalue == nil{
				continue
			}
			switch recordvalue.DataType {
			case dxsvalue.VT_String,dxsvalue.VT_RawString:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(recordvalue.String()))
			case dxsvalue.VT_Int:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(recordvalue.Int()))
			case dxsvalue.VT_False:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(false))
			case dxsvalue.VT_True:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(true))
			case dxsvalue.VT_DateTime:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(recordvalue.GoTime()))
			case dxsvalue.VT_Double:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(recordvalue.Double()))
			case dxsvalue.VT_Float:
				v.SetMapIndex(reflect.ValueOf(k),reflect.ValueOf(recordvalue.Float()))
			}
		}
		return nil
	}
	return nil
}

type redisJsonCoder struct {
	v 				reflect.Value
	valueType		*redisFields
	bf				[]byte
}

var(
	jsonCoderPool		sync.Pool
)

func getredisJsonCoder(v reflect.Value,valueType *redisFields )*redisJsonCoder  {
	var result *redisJsonCoder
	value := jsonCoderPool.Get()
	if value == nil{
		result = &redisJsonCoder{
			v: v,
			valueType: valueType,
			bf: make([]byte,0,512),
		}
	}else{
		result = value.(*redisJsonCoder)
		result.bf = result.bf[:0]
	}
	return result
}

func (coder *redisJsonCoder)writeObject(v reflect.Value,dst []byte,tp reflect.Type)[]byte  {
	valueFields := fieldsByType(tp)
	dst = append(dst,'{')
	isfirst := true
	writeKeyName := func(keyName string) {
		if !isfirst{
			dst = append(dst,`,"`...)
		}else{
			dst = append(dst,'"')
			isfirst = false
		}
		dst = append(dst,DxCommonLib.FastString2Byte(keyName)...)
		dst = append(dst,`":`...)
	}

	for i := 0;i<len(valueFields.fields);i++{
		if valueFields.fields[i].isKey{
			continue
		}
		f := v.Field(int(valueFields.fields[i].index))
		switch valueFields.fields[i].fieldType.Kind() {
		case reflect.String:
			writeKeyName(valueFields.fields[i].fieldName)
			dst = append(dst,'"')
			dst = append(dst,DxCommonLib.FastString2Byte(f.String())...)
			dst = append(dst,'"')
		case reflect.Bool:
			writeKeyName(valueFields.fields[i].fieldName)
			if f.Bool(){
				dst = append(dst,"true"...)
			}else{
				dst = append(dst,"false"...)
			}
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			writeKeyName(valueFields.fields[i].fieldName)
			dst = append(dst,strconv.Itoa(int(f.Int()))...)
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			writeKeyName(valueFields.fields[i].fieldName)
			dst = append(dst,strconv.FormatUint(f.Uint(),10)...)
		case reflect.Float32,reflect.Float64:
			writeKeyName(valueFields.fields[i].fieldName)
			dst = append(dst,strconv.FormatFloat(f.Float(),'f',-1,64)...)
		case reflect.Struct:
			if valueFields.fields[i].fieldType == TimeType{
				writeKeyName(valueFields.fields[i].fieldName)
				dst = append(dst,"\"/Date("...)
				unixs := int64(f.Interface().(time.Time).Unix()*1000)
				dst = strconv.AppendInt(dst,unixs,10)
				dst = append(dst,")/\""...)
			}else if f.Type().Name() == valueFields.fields[i].fieldName{
				//内嵌结构体，解析出这个内嵌结构的内容
				writeKeyName(valueFields.fields[i].fieldName)
				dst = coder.writeObject(f,dst,valueFields.fields[i].fieldType)
			}
		}
	}
	dst = append(dst,'}')
	return dst
}

func (coder *redisJsonCoder)MarshalJSON() ([]byte, error)  {
	dst := coder.bf
	isfirst := true
	dst = append(dst,'{')
	writeKeyName := func(keyName string) {
		if !isfirst{
			dst = append(dst,`,"`...)
		}else{
			dst = append(dst,'"')
			isfirst = false
		}
		dst = append(dst,DxCommonLib.FastString2Byte(keyName)...)
		dst = append(dst,`":`...)
	}
	for i := 0;i<len(coder.valueType.fields);i++{
		if coder.valueType.fields[i].isKey{
			if !coder.valueType.fields[i].IsInnerStruct(){ //内嵌结构体，还要判定一下
				continue
			}
		}
		f := coder.v.Field(int(coder.valueType.fields[i].index))
		tp := coder.valueType.fields[i].fieldType
		switch tp.Kind() {
		case reflect.String:
			writeKeyName(coder.valueType.fields[i].fieldName)
			dst = append(dst,'"')
			dst = append(dst,DxCommonLib.FastString2Byte(f.String())...)
			dst = append(dst,'"')
		case reflect.Bool:
			writeKeyName(coder.valueType.fields[i].fieldName)
			if f.Bool(){
				dst = append(dst,"true"...)
			}else{
				dst = append(dst,"false"...)
			}
		case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
			writeKeyName(coder.valueType.fields[i].fieldName)
			dst = append(dst,strconv.Itoa(int(f.Int()))...)
		case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
			writeKeyName(coder.valueType.fields[i].fieldName)
			dst = append(dst,strconv.FormatUint(f.Uint(),10)...)
		case reflect.Float32,reflect.Float64:
			writeKeyName(coder.valueType.fields[i].fieldName)
			dst = append(dst,strconv.FormatFloat(f.Float(),'f',-1,64)...)
		case reflect.Struct:
			if tp == TimeType{
				writeKeyName(coder.valueType.fields[i].fieldName)
				dst = append(dst,"\"/Date("...)
				unixs := int64(f.Interface().(time.Time).Unix()*1000)
				dst = strconv.AppendInt(dst,unixs,10)
				dst = append(dst,")/\""...)
			}else if coder.valueType.fields[i].IsInnerStruct(){
				//内嵌结构体，解析出这个内嵌结构的内容
				writeKeyName(coder.valueType.fields[i].fieldName)
				dst = coder.writeObject(f,dst,tp)
			}
		}
	}
	dst = append(dst,'}')
	return dst,nil
}

func encodeRedis(v reflect.Value,valuetype *redisFields,vredis map[string]interface{})error  {
	if valuetype == nil {
		return errors.New("无法Encode到Redis")
	}
	if valuetype != nil && len(valuetype.fields) == 0{
		return nil
	}
	//构建redis的hashKey
	hashkey := valuetype.hashKey(v)
	if hashkey == ""{
		return nil
	}
	//将redis指定的Key信息过滤掉
	jsoncoder := getredisJsonCoder(v,valuetype)
	bt,err := json.Marshal(jsoncoder)
	if err != nil{
		jsoncoder.valueType = nil
		jsonCoderPool.Put(jsoncoder)
		return err
	}
	vredis[hashkey] = string(bt)
	jsoncoder.valueType = nil
	jsonCoderPool.Put(jsoncoder)
	return nil
}

func getFromRedisMap(resultmap map[string]string,reflectv reflect.Value,valueType reflect.Type)error  {
	n := len(resultmap)
	if n == 0{
		return nil
	}
	if reflectv.Cap() >= n {
		reflectv.Set(reflectv.Slice(0, n))
	} else if reflectv.Len() < reflectv.Cap() {
		reflectv.Set(reflectv.Slice(0, reflectv.Cap()))
	}
	var valueFields *redisFields

	if valueType.Kind() == reflect.Struct{
		valueFields = fieldsByType(valueType)
	}
	//是否自己实现了编码解码功能
	if valueType.Implements(RedisValueCoderType){
		idx := 0
		for VKey,vredis := range resultmap{
			if idx >= reflectv.Len() {
				reflectv.Set(growSliceValue(reflectv, n))
			}
			//解析redis
			sv := reflectv.Index(idx)
			idx ++
			err := sv.Interface().(RedisValueCoder).Decode(VKey,vredis)
			if err != nil{
				return err
			}
		}
	}else{
		idx := 0
		for VKey,vredis := range resultmap {
			if idx >= reflectv.Len() {
				reflectv.Set(growSliceValue(reflectv, n))
			}
			//解析redis
			sv := reflectv.Index(idx)
			idx++
			//解析值

			err := decodeRedis(&sv,VKey,vredis,valueFields)
			if err != nil{
				return err
			}
		}
	}
	return nil
}

//返回的是KV数组，两个作为一组
func getFromRedisSlice(resultSlice []string,reflectv reflect.Value,valueType reflect.Type,valueFields *redisFields)error  {
	n := len(resultSlice)
	if n == 0{
		return nil
	}
	if reflectv.Cap() >= n {
		reflectv.Set(reflectv.Slice(0, n))
	} else if reflectv.Len() < reflectv.Cap() {
		reflectv.Set(reflectv.Slice(0, reflectv.Cap()))
	}
	//是否自己实现了编码解码功能
	if valueType.Implements(RedisValueCoderType){
		idx := 0
		for i := 0;i<len(resultSlice);i += 2{
			key := resultSlice[i]
			value := resultSlice[i+1]
			if idx >= reflectv.Len() {
				reflectv.Set(growSliceValue(reflectv, n))
			}
			//解析redis
			sv := reflectv.Index(idx)
			idx ++
			vcoder := sv.Interface().(RedisValueCoder)
			err := vcoder.Decode(key,value)
			if err != nil{
				return err
			}
		}
	}else{
		idx := 0
		for i := 0;i<len(resultSlice);i += 2{
			if idx >= reflectv.Len() {
				reflectv.Set(growSliceValue(reflectv, n))
			}
			//解析redis
			sv := reflectv.Index(idx)
			idx++
			//解析值
			key := resultSlice[i]
			value := resultSlice[i+1]
			err := decodeRedis(&sv,key,value,valueFields)
			if err != nil{
				return err
			}
		}
	}
	return nil
}


//查询指定的KV条件
func Find(dest interface{},queryInfo map[string]string)error  {
	reflectv := reflect.ValueOf(dest)
	if !reflectv.IsValid(){
		panic("dest Invalidate")
	}
	if reflectv.Kind() != reflect.Ptr{
		panic("dest must pointer")
	}
	reflectv = reflectv.Elem()
	vt := reflectv.Type()
	if vt.Kind() != reflect.Slice{
		panic("dest type must Slice")
	}
	vslice := reflect.MakeSlice(vt, 1, 1)
	sv := vslice.Index(0)
	valueType := sv.Type()
	if valueType.Kind() != reflect.Struct && !valueType.Implements(RedisValueCoderType){
		panic("dest value must be struct or implements RedisValueCoder interface")
	}
	//获取redisKey
	redisKey := ""
	if valueType.Implements(RedisTableInter){
		redisKey = sv.Interface().(RedisTable).TableName()
	}else{
		redisKey = valueType.Name()
	}
	redisKey = "RedisDB:"+redisKey
	if queryInfo == nil || len(queryInfo) == 0{
		resultMap,err := HGetAll(context.Background(),redisKey).Result()
		if err != nil{
			return err
		}
		return getFromRedisMap(resultMap,reflectv,valueType)
	}

	var valueFields *redisFields
	valueFields = fieldsByType(valueType)
	//首先要将where条件根据索引key序号排序
	querKey := valueFields.CreateIndexKey(queryInfo)
	resultKvs,_,err := HScan(context.Background(),redisKey,0,querKey,math.MaxUint16).Result()
	if err != nil{
		return err
	}
	//返回的就是KV结构
	return getFromRedisSlice(resultKvs,reflectv,valueType,valueFields)
}

//API:表名   Key1:Key2:Key3    内容是Json内容{}
func SaveTable(dest interface{})(bool,error)  {
	reflectv := reflect.ValueOf(dest)
	if !reflectv.IsValid(){
		panic("dest Invalidate")
	}
	if reflectv.Kind() == reflect.Ptr{
		reflectv = reflectv.Elem()
	}
	vt := reflectv.Type()
	if vt.Kind() != reflect.Slice{
		panic("dest type must Slice")
	}
	return updateSlice(reflectv)
}

func updateSlice(reflectv reflect.Value)(bool,error)  {
	if reflectv.Len() == 0{
		return true,nil
	}
	sv := reflectv.Index(0)
	valueType := sv.Type()
	redisKey := ""
	//获取redisKey
	if !valueType.Implements(RedisTableInter){
		if valueType.Kind() == reflect.Struct{
			redisKey = valueType.Name()
		}else{
			panic("dest must be RedisTable")
		}
	}else{
		redisKey = sv.Interface().(RedisTable).TableName()
	}
	redisKey = "RedisDB:"+redisKey
	//是否自己实现了编码解码功能
	vredis := make(map[string]interface{},reflectv.Len())
	if valueType.Implements(RedisValueCoderType){
		for i := 0;i<reflectv.Len();i++{
			rv := reflectv.Index(i)
			valueEncoder := rv.Interface().(RedisValueCoder)
			vredis[valueEncoder.EncodeHash()] = valueEncoder.Encode()
		}
	}else{
		valueFields := fieldsByType(valueType)
		for i := 0;i<reflectv.Len();i++ {
			rv := reflectv.Index(i)
			err := encodeRedis(rv,valueFields,vredis)
			if err != nil{
				return false, err
			}
		}
	}
	return HMSet(context.Background(),redisKey,vredis).Result()

}

//更新一条或多条数据
func Update(value interface{})(bool,error)  {
	reflectv := reflect.ValueOf(value)
	if !reflectv.IsValid(){
		panic("dest Invalidate")
	}
	if reflectv.Kind() == reflect.Ptr{
		reflectv = reflectv.Elem()
	}
	switch reflectv.Kind() {
	case reflect.Slice:
		return updateSlice(reflectv)
	case reflect.Struct:
		valueType := reflectv.Type()
		redisKey := ""
		if !valueType.Implements(RedisTableInter){
			if valueType.Kind() == reflect.Struct{
				redisKey = valueType.Name()
			}else{
				panic("dest must be RedisTable")
			}
		}else{
			redisKey = reflectv.Interface().(RedisTable).TableName()
		}
		redisKey = "RedisDB:"+ redisKey
		if valueType.Implements(RedisValueCoderType){
			valueEncoder := reflectv.Interface().(RedisValueCoder)
			_,err := HSet(context.Background(),redisKey,valueEncoder.EncodeHash(),valueEncoder.Encode()).Result()
			return err == nil,err
		}
		valueFields := fieldsByType(valueType)
		hashkey := valueFields.hashKey(reflectv)
		jsoncoder := getredisJsonCoder(reflectv,valueFields)
		bt,err := json.Marshal(jsoncoder)
		if err != nil{
			jsoncoder.valueType = nil
			jsonCoderPool.Put(jsoncoder)
			return false,err
		}
		v := string(bt)
		jsoncoder.valueType = nil
		jsonCoderPool.Put(jsoncoder)
		_,err = HSet(context.Background(),redisKey,hashkey,v).Result()
		return err == nil,err
	default:
		panic("不支持的结构")
	}
	return false,nil
}

func deleteslice(reflectv reflect.Value)(int64,error)  {
	if reflectv.Len() == 0{
		return 0,nil
	}
	sv := reflectv.Index(0)
	valueType := sv.Type()
	if valueType.Kind() != reflect.Struct && !valueType.Implements(RedisValueCoderType){
		panic("dest value must be struct or implements RedisValueCoder interface")
	}
	redisKey := ""
	if valueType.Implements(RedisTableInter){
		redisKey = sv.Interface().(RedisTable).TableName()
	}else{
		redisKey = valueType.Name()
	}
	redisKey = "RedisDB:"+redisKey
	hashKey := make([]string,0,4)
	if valueType.Implements(RedisValueCoderType){
		for i := 0;i<reflectv.Len();i++{
			sv = reflectv.Index(i)
			hashstr := sv.Interface().(RedisValueCoder).EncodeHash()
			if hashstr != ""{
				hashKey = append(hashKey,hashstr)
			}
		}
	}else{
		valueFields := fieldsByType(valueType)
		for i := 0;i<reflectv.Len();i++{
			hashstr := valueFields.hashKey(reflectv.Index(i))
			if hashstr != ""{
				hashKey = append(hashKey,hashstr)
			}
		}
	}
    return HDel(context.Background(),redisKey,hashKey...).Result()
}

//删除数据,可以是结构，也可以是slice
func Delete(value interface{})(int64,error)  {
	reflectv := reflect.ValueOf(value)
	if !reflectv.IsValid(){
		panic("dest Invalidate")
	}
	if reflectv.Kind() == reflect.Ptr{
		reflectv = reflectv.Elem()
	}
	switch reflectv.Kind() {
	case reflect.Slice:
		return deleteslice(reflectv)
	case reflect.Struct:
		valueType := reflectv.Type()
		hashStr := ""
		redisKey := ""
		if valueType.Implements(RedisTableInter){
			redisKey = reflectv.Interface().(RedisTable).TableName()
		}else{
			redisKey = valueType.Name()
		}
		redisKey = "RedisDB:"+redisKey
		if valueType.Implements(RedisValueCoderType){
			hashStr = reflectv.Interface().(RedisValueCoder).EncodeHash()
		}else{
			valuefields := fieldsByType(valueType)
			hashStr = valuefields.hashKey(reflectv)
		}
		if hashStr != ""{
			return HDel(context.Background(),redisKey,hashStr).Result()
		}
	case reflect.String:
		//删除某个hashmap
		Del(context.Background(),"RedisDB:"+reflectv.String()).Result()
	default:
		panic("不支持的结构")
	}
	return 0,nil
}

func Drop(tableName string)(int64,error) {
	return Del(context.Background(),"RedisDB:"+tableName).Result()
}

//dest中必须标记好主键信息,直接调用HGet返回
func First(dest interface{})(bool,error)  {
	reflectv := reflect.ValueOf(dest)
	if !reflectv.IsValid(){
		panic("dest Invalidate")
	}
	if reflectv.Kind() != reflect.Ptr{
		panic("dest must pointer")
	}
	reflectv = reflectv.Elem()
	valueType := reflectv.Type()
	redisKey := ""
	if valueType.Implements(RedisTableInter){
		redisKey = reflectv.Interface().(RedisTable).TableName()
	}else if valueType.Kind() == reflect.Struct{
		redisKey = valueType.Name()
	}
	if redisKey == ""{
		return false,nil
	}
	redisKey = "RedisDB:"+redisKey
	hashKey := ""
	var valueCoder RedisValueCoder
	var valueFields *redisFields
	if valueType.Implements(RedisValueCoderType){
		valueCoder = reflectv.Interface().(RedisValueCoder)
		if hashKey == ""{
			hashKey = valueCoder.EncodeHash()
		}
	}else{
		if valueType.Kind() != reflect.Struct{
			panic("不支持的")
		}
		valueFields = fieldsByType(valueType)
		mp := make(map[string]string,4)
		valueFields.hashKeymap(reflectv,mp)
		hashKey = valueFields.CreateIndexKey(mp)
	}
	results,_,err := HScan(context.Background(),redisKey,0,hashKey,1).Result()
	if err != nil{
		return false,err
	}
	for i := 0;i<len(results);i += 2{
		key := results[i]
		value := results[i+1]
		if valueCoder != nil{
			err = valueCoder.Decode(key,value)
		}else{
			err = decodeRedis(&reflectv,key,value,valueFields)
		}
		break
	}
	return err == nil,err
}

