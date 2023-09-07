package orm

import (
	"common/log"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"regexp"
	"time"
)

type Logger struct {
	LogLevel logger.LogLevel
}

// LogMode log mode
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l *Logger) Info(_ context.Context, sql string, params ...interface{}) {
	log.Log(5, logrus.InfoLevel, sql, params)
}
func (l *Logger) Warn(_ context.Context, sql string, params ...interface{}) {
	log.Log(5, logrus.WarnLevel, sql, params)
}
func (l *Logger) Error(_ context.Context, sql string, params ...interface{}) {
	log.Log(5, logrus.ErrorLevel, sql, params)
}

func (l Logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rowsAffected := fc()
	sql = regexp.MustCompile("\\s+").ReplaceAllString(sql, " ")
	if time.Now().After(begin.Add(time.Second * 3)) {
		log.Log(5, logrus.WarnLevel, fmt.Sprintf("<%v> slow sql statement: %s, rowsAffected: %d",
			time.Since(begin), sql, rowsAffected))
		return
	}
	if l.LogLevel == logger.Info && err == nil {
		log.Log(5, logrus.DebugLevel, fmt.Sprintf("<%v> sql statement: %s, rowsAffected: %d",
			time.Since(begin), sql, rowsAffected))
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Log(5, logrus.WarnLevel, fmt.Sprintf("<%v> sql statement no record: %s",
			time.Since(begin), sql))
		return
	}
	if err != nil {
		log.Log(5, logrus.ErrorLevel, fmt.Sprintf("bad sql: %s, error: %v", sql, err))
	}
}
