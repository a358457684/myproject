package util

import (
	uuid "github.com/satori/go.uuid"
	"strconv"
	"strings"
	"time"
)

func CreateNormalUUID() string {
	return uuid.NewV4().String()
}

func CreateUUID() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

func GetDel() string {
	return "|del=" + strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
}
