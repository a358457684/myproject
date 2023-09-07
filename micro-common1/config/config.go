package config

import (
	"fmt"
	"github.com/suiyunonghen/dxsvalue"
	"reflect"
	"time"
)

var (
	Data               Options
	CurrentConfigPath  string
	configFile         string //编译时指定配置文件  go build -ldflags "-X common/config.configFile=conf/config-test.yml"
	defaultConfigFiles = []string{"./conf/config.yml", "../conf/config.yml", "../../conf/config.yml",
		"./config.yml", "../config.yml", "../../config.yml"}
)

func string2Duration(fvalue reflect.Value, value *dxsvalue.DxValue) {
	switch value.DataType {
	case dxsvalue.VT_Int, dxsvalue.VT_Float, dxsvalue.VT_Double:
		fvalue.SetInt(value.Int())
	case dxsvalue.VT_DateTime:
		fvalue.SetInt(int64(time.Now().Sub(value.GoTime())))
	case dxsvalue.VT_String:
		duration, err := time.ParseDuration(value.String())
		if err == nil {
			fvalue.SetInt(int64(duration))
		}
	}
}

func init() {
	//注册一下类型
	TimeDurationPtrType := reflect.TypeOf((*time.Duration)(nil))
	TimeDurationType := TimeDurationPtrType.Elem()
	dxsvalue.RegisterTypeMapFunc(TimeDurationType, string2Duration)
	dxsvalue.RegisterTypeMapFunc(TimeDurationPtrType, string2Duration)
	Data.Log = DefaultLogOptions()
	if configFile != "" {
		defaultConfigFiles = []string{configFile}
	}
	l := len(defaultConfigFiles)
	for i := 0; i < len(defaultConfigFiles); i++ {
		value, err := dxsvalue.NewValueFromYamlFile(defaultConfigFiles[i], true)
		if err != nil {
			if i == l-1 {
				fmt.Printf("配置文件加载失败，%v", err)
			}
			continue
		}
		Data.LoadFromValue(value)
		CurrentConfigPath = defaultConfigFiles[i]
		dxsvalue.FreeValue(value)
		fmt.Printf("配置文件加载完成，%v\n", CurrentConfigPath)
		return
	}
}
