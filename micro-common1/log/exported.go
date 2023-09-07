package log

import (
	"common/config"
	"context"
	"github.com/sirupsen/logrus"
)

var std *Logger

func DefLogger() *Logger {
	return std
}

func init() {
	logOptions := config.Data.Log
	if logOptions == nil {
		logOptions = config.DefaultLogOptions()
	}
	std = NewLog(logOptions)
	std.Info("日志系统初始化成功")
}

func Init(options *config.LogOptions) {
	std = NewLog(options)
	std.Info("日志系统自定义初始化成功")
}

// WithError creates an entry from the standard log and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *logrus.Entry {
	return std.withError(1, err)
}

// WithContext creates an entry from the standard log and adds a context to it.
func WithContext(ctx context.Context) *logrus.Entry {
	return std.withContext(1, ctx)
}

// WithField creates an entry from the standard log and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithField(key string, value interface{}) *logrus.Entry {
	return std.withField(1, key, value)
}

// WithFields creates an entry from the standard log and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
func WithFields(fields logrus.Fields) *logrus.Entry {
	return std.withFields(1, fields)
}

func WithOffsetFields(offset int, fields logrus.Fields) *logrus.Entry {
	return std.withFields(offset, fields)
}

// Trace logs a message at Level Trace on the standard log.
func Trace(args ...interface{}) {
	std.log(2, logrus.TraceLevel, args...)
}

// Debug logs a message at Level Debug on the standard log.
func Debug(args ...interface{}) {
	std.log(2, logrus.DebugLevel, args...)
}

// Print logs a message at Level Info on the standard log.
func Print(args ...interface{}) {
	std.log(2, logrus.InfoLevel, args...)
}

// Info logs a message at Level Info on the standard log.
func Info(args ...interface{}) {
	std.log(2, logrus.InfoLevel, args...)
}

// Warn logs a message at Level Warn on the standard log.
func Warn(args ...interface{}) {
	std.log(2, logrus.WarnLevel, args...)
}

// Warning logs a message at Level Warn on the standard log.
func Warning(args ...interface{}) {
	std.log(2, logrus.WarnLevel, args...)
}

// Error logs a message at Level Error on the standard log.
func Error(args ...interface{}) {
	std.log(2, logrus.ErrorLevel, args...)
}

// Panic logs a message at Level Panic on the standard log.
func Panic(args ...interface{}) {
	std.log(2, logrus.PanicLevel, args...)
}

// Fatal logs a message at Level Fatal on the standard log then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	std.log(2, logrus.FatalLevel, args...)
	std.l.Exit(1)
}

// Tracef logs a message at Level Trace on the standard log.
func Tracef(format string, args ...interface{}) {
	std.logf(2, logrus.TraceLevel, format, args...)
}

// Debugf logs a message at Level Debug on the standard log.
func Debugf(format string, args ...interface{}) {
	std.logf(2, logrus.DebugLevel, format, args...)
}

// Printf logs a message at Level Info on the standard log.
func Printf(format string, args ...interface{}) {
	std.logf(2, logrus.InfoLevel, format, args...)
}

// Infof logs a message at Level Info on the standard log.
func Infof(format string, args ...interface{}) {
	std.logf(2, logrus.InfoLevel, format, args...)
}

// Warnf logs a message at Level Warn on the standard log.
func Warnf(format string, args ...interface{}) {
	std.logf(2, logrus.WarnLevel, format, args...)
}

// Warningf logs a message at Level Warn on the standard log.
func Warningf(format string, args ...interface{}) {
	std.logf(2, logrus.WarnLevel, format, args...)
}

// Errorf logs a message at Level Error on the standard log.
func Errorf(format string, args ...interface{}) {
	std.logf(2, logrus.ErrorLevel, format, args...)
}

// Panicf logs a message at Level Panic on the standard log.
func Panicf(format string, args ...interface{}) {
	std.logf(2, logrus.PanicLevel, format, args...)
}

// Fatalf logs a message at Level Fatal on the standard log then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	std.logf(2, logrus.FatalLevel, format, args...)
	std.l.Exit(1)
}

// Traceln logs a message at Level Trace on the standard log.
func Traceln(args ...interface{}) {
	std.logln(2, logrus.TraceLevel, args...)
}

// Debugln logs a message at Level Debug on the standard log.
func Debugln(args ...interface{}) {
	std.logln(2, logrus.DebugLevel, args...)
}

// Println logs a message at Level Info on the standard log.
func Println(args ...interface{}) {
	std.logln(2, logrus.InfoLevel, args...)
}

// Infoln logs a message at Level Info on the standard log.
func Infoln(args ...interface{}) {
	std.logln(2, logrus.InfoLevel, args...)
}

// Warnln logs a message at Level Warn on the standard log.
func Warnln(args ...interface{}) {
	std.logln(2, logrus.WarnLevel, args...)
}

// Warningln logs a message at Level Warn on the standard log.
func Warningln(args ...interface{}) {
	std.logln(2, logrus.WarnLevel, args...)
}

// Errorln logs a message at Level Error on the standard log.
func Errorln(args ...interface{}) {
	std.logln(2, logrus.ErrorLevel, args...)
}

// Panicln logs a message at Level Panic on the standard log.
func Panicln(args ...interface{}) {
	std.logln(2, logrus.PanicLevel, args...)
}

// Fatalln logs a message at Level Fatal on the standard log then the process will exit with status set to 1.
func Fatalln(args ...interface{}) {
	std.logln(2, logrus.FatalLevel, args...)
	std.l.Exit(1)
}

func Log(offset int, level logrus.Level, args ...interface{}) {
	std.log(offset, level, args...)
}

// err不为空才进行记录
func LogWithError(err error, some ...interface{}) {
	if err != nil {
		std.withError(1, err).Error(some...)
	}
}

func LogWithErrorf(err error, format string, some ...interface{}) {
	if err != nil {
		std.withError(1, err).Errorf(format, some...)
	}
}

// err为空时使用info记录
func MustLogWithError(err error, some ...interface{}) {
	if err != nil {
		std.withError(1, err).Error(some...)
	} else {
		Info(some...)
	}
}

func MustLogWithErrorf(err error, format string, some ...interface{}) {
	if err != nil {
		std.WithError(err).Errorf(format, some...)
	} else {
		std.Infof(format, some...)
	}
}
