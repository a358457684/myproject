// +build json // 仅当编译时带 -tags=json 参数时才生效

package jsonutil

import (
	"encoding/json"
	"strings"
)

var (
	Marshal         = json.Marshal
	Unmarshal       = json.Unmarshal
	MarshalToString = func(v interface{}) (string, error) {
		data, err := json.Marshal(v)
		if nil != err {
			return "", err
		}
		return string(data), nil
	}
	UnmarshalFromString = func(str string, v interface{}) error {
		str = strings.TrimSpace(str)
		if str == "" {
			return nil
		}

		data := []byte(str)
		return json.Unmarshal(data, v)
	}
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)
