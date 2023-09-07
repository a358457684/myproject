package mq

import "pp/common-golang/date"

type TestMsg struct {
	Message string        `json:"message"`
	Time    date.Datetime `json:"time"`
}
