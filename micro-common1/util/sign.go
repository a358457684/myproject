package util

import (
	"sort"
)

func SortParam(paramsMap map[string]string, securityKey string) string {
	sign := ""
	params := make([]string, len(paramsMap))
	for k := range paramsMap {
		params = append(params, k)
	}
	params = sortstr(params)

	num := 0
	for i := 0; i < len(params); i++ {
		str := params[i]
		if paramsMap[str] != "" {
			if num > 0 {
				sign = sign + "&"
			}
			sign += str + "=" + paramsMap[str]
			num++
		}
	}

	sign += "&securityKey=" + securityKey
	return sign
}

func sortstr(params []string) []string {
	sort.Slice(params, func(i, j int) bool {

		for m := 0; m < len(params[i]) && m < len(params[j]); m++ {
			if params[i][m] == params[j][m] {
				continue
			}
			if params[i][m] > params[j][m] {
				return false
			} else {

				return true
			}
		}
		return true
	})
	return params
}
