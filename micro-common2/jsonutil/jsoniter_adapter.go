// 设置jsoniter为默认库 // +build jsoniter // 仅当编译时带 -tags=jsoniter 参数时才生效

package jsonutil

import (
	"github.com/json-iterator/go"
)

var (
	adapter             = jsoniter.ConfigCompatibleWithStandardLibrary
	Marshal             = adapter.Marshal
	Unmarshal           = adapter.Unmarshal
	MarshalToString     = adapter.MarshalToString
	UnmarshalFromString = adapter.UnmarshalFromString
	MarshalIndent       = adapter.MarshalIndent
	NewDecoder          = adapter.NewDecoder
	NewEncoder          = adapter.NewEncoder
)
