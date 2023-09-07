package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

type Config struct {
	SlowThreshold time.Duration
	Colorful      bool
}

type Log struct {
	logEntity *Logger
	Config    Config

	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func InitGormLog(log *Logger, config Config) logger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}

	l := &Log{
		logEntity:    log,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}

	return l
}

func (l *Log) LogMode(level logger.LogLevel) logger.Interface {
	switch level {
	case logger.Silent:
		l.logEntity.Level = logrus.InfoLevel
	case logger.Error:
		l.logEntity.Level = logrus.ErrorLevel
	case logger.Warn:
		l.logEntity.Level = logrus.WarnLevel
	case logger.Info:
		l.logEntity.Level = logrus.InfoLevel
	}
	return l
}

func (l *Log) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logEntity.Level >= logrus.InfoLevel {
		l.logEntity.WithContext(ctx).Infof(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Log) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logEntity.WithContext(ctx).Warnf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l *Log) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logEntity.WithContext(ctx).Errorf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l *Log) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logEntity.Level > 0 {
		elapsed := time.Since(begin)
		switch {
		case err != nil && l.logEntity.Level >= logrus.ErrorLevel:
			sql, rows := fc()
			if rows == -1 {
				l.logEntity.WithContext(ctx).Errorf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.logEntity.WithContext(ctx).Errorf(l.traceErrStr, utils.FileWithLineNum(), err,
					float64(elapsed.Nanoseconds())/1e6,
					rows, sql)
			}
		case elapsed > l.Config.SlowThreshold && l.Config.SlowThreshold != 0 && l.logEntity.Level >= logrus.WarnLevel:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.Config.SlowThreshold)
			if rows == -1 {
				l.logEntity.WithContext(ctx).Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.logEntity.WithContext(ctx).Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case l.logEntity.Level >= logrus.InfoLevel:
			sql, rows := fc()
			if rows == -1 {
				l.logEntity.WithContext(ctx).Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.logEntity.WithContext(ctx).Infof(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}
