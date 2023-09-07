package log

import (
	"common/config"
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	Info("测试内容\r\n换行")
	WithField("key", "value").Info("测试")
	WithError(errors.New("异常")).Error("错误")
}

func TestLogger_Customer_WithFields(t *testing.T) {
	logOptions := config.DefaultLogOptions()
	logOptions.Project = "test"
	logOptions.Author = "不得闲"
	logOptions.Level = "debug"
	logOptions.File = "logs/test.log"
	logOptions.SplitTime = 0
	logOptions.SplitSize = 6 * 1024
	Init(logOptions)
	for i := 1000; i < 2000; i++ {
		WithField("ID", strconv.Itoa(i)).Info("Asdfasdf")
		time.Sleep(time.Millisecond * 50)
	}
}

func TestLogger_Customer_LazyWrite(t *testing.T) {
	logOptions := config.DefaultLogOptions()
	logOptions.Project = "工程"
	logOptions.Author = "作者"
	logOptions.LazyWrite = true
	logOptions.Level = "debug"
	logOptions.File = "logs/test.log"
	Init(logOptions)
	WithField("ff", 3).Debug("Asdfasdf")
	time.Sleep(time.Second * 7)
}
