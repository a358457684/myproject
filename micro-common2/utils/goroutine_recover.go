package utils

import (
	"log"
	"pp/common-golang/logger"
)

func DefaultGoroutineRecover(l *logger.Logger, action string) {
	if err := recover(); err != nil {
		if e, ok := err.(error); ok {
			if nil != l {
				l.WithField("err", e.Error()).Error(action, " goroutine 异常")
			} else {
				log.Print(action, " goroutine 异常 ", e.Error())
			}
			stack := string(GetStack(5))
			println(stack)
		}
	}
}
