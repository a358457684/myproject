package util

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var commonAbbr = []string{"ID", "API"}
var titleReplace *strings.Replacer
var upReplace *strings.Replacer

func init() {
	var lowerForReplacer []string
	var upForReplacer []string
	for _, abbr := range commonAbbr {
		title := strings.Title(strings.ToLower(abbr))
		lowerForReplacer = append(lowerForReplacer, abbr, title)
		upForReplacer = append(upForReplacer, title, abbr)
	}
	titleReplace = strings.NewReplacer(lowerForReplacer...)
	upReplace = strings.NewReplacer(upForReplacer...)
}

// 驼峰式写法转为下划线写法
func UnderscoreName(name string) string {
	name = titleReplace.Replace(name)
	buffer := bytes.Buffer{}
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.WriteByte('_')
			}
			buffer.WriteRune(unicode.ToLower(r))
		} else {
			buffer.WriteRune(r)
		}
	}
	return buffer.String()
}

// 下划线写法转为驼峰写法
func CamelName(name string) string {
	name = strings.Replace(name, "_", " ", -1)
	name = upReplace.Replace(strings.Title(name))
	return strings.Replace(name, " ", "", -1)
}

func Rmspec(source string, args ...uint8) string {
	if source == "" {
		return source
	}
	start, end := 0, 0
	for i := 0; i < len(source); i++ {
		flag := 0
		for _, item := range args {
			if source[i] == item {
				flag = 1
				break
			}
		}
		if flag == 0 {
			start = i
			goto Endloop
		}
	}

Endloop:
	for j := len(source) - 1; j >= 0; j-- {
		flag := 0
		for _, item := range args {
			if source[j] == item {
				flag = 1
				break
			}
		}
		if flag == 0 {
			end = j
			goto End
		}
	}
End:

	if end >= start {
		return source[start : end+1]
	} else {
		return ""
	}
}

func Rmbrackets(source string) string {
	return Rmspec(source, 91, 93)
}

func EndSplitTwo(val, split string) (string, string) {
	idx := strings.LastIndex(val, split)
	if idx < 0 {
		return val, ""
	} else {
		return val[:idx], val[idx+1:]
	}
}

func ParamsCheck(args ...string) bool {
	for _, item := range args {
		if item == "" {
			return false
		}
	}
	return true
}

func IsNotEqualsAndNotEmpty(pid, id string) bool {
	if pid == "" || id == "" {
		return false
	}
	return pid != id
}

func ConstainKey(mp map[string]interface{}, key string) bool {
	if _, ok := mp[key]; ok {
		return true
	}
	return false
}

func Isnotblank(str string) bool {
	return !Isblank(str)
}

func Isblank(str string) bool {
	if len(str) > 0 {
		for _, item := range str {
			if item != 32 && item != 9 {
				return false
			}
		}
	}
	return true
}

func IsblankSpec(str string) interface{} {
	if len(str) > 0 {
		for _, item := range str {
			if item != 32 && item != 9 {
				return false
			}
		}
	}
	return true
}

func IsNumber(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

func GetIndex(preindex string, indexDate time.Time) string {
	var formatDateStr string
	if (time.Time{}) != indexDate {
		formatDateStr = time.Unix(indexDate.Unix(), 0).Format("2006.01")
	} else {
		formatDateStr = time.Unix(time.Now().Unix(), 0).Format("2006.01")
	}
	index := preindex + formatDateStr
	return index
}

func CompareTo(src, anotherString string) int {
	len1 := len(src)
	len2 := len(anotherString)
	lim := int(math.Min(float64(len1), float64(len2)))
	v1 := []byte(src)
	v2 := []byte(anotherString)

	k := 0
	for {
		if k < lim {
			c1 := v1[k]
			c2 := v2[k]
			if c1 != c2 {
				return int(c1) - int(c2)
			}
			k++
		} else {
			break
		}
	}

	return len1 - len2
}
