package util

import (
	uuid "github.com/satori/go.uuid"
	"strings"
)

func GetUUID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}
