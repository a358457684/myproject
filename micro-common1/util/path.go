package util

import "strings"

func FormatPath(path string) string {
	if strings.LastIndex(path, "/") == len(path)-1 {
		return path
	}
	return path + "/"
}
