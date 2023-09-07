package logger

import (
	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

type XORMLogger struct {
	logger  *Logger
	level   log.LogLevel
	showSQL bool
}

func NewXORMLogger(logger *Logger) log.Logger {
	lg := &XORMLogger{logger: logger.WithField("lib", "xorm"), showSQL: true}
	level := log.LOG_INFO
	if logger.Level == logrus.WarnLevel {
		level = log.LOG_WARNING
	} else if logger.Level == logrus.DebugLevel {
		level = log.LOG_DEBUG
	} else if logger.Level == logrus.ErrorLevel {
		level = log.LOG_ERR
	} else if logger.Level == logrus.FatalLevel {
		level = log.LOG_OFF
	} else if logger.Level == logrus.PanicLevel {
		level = log.LOG_UNKNOWN
	}
	lg.level = level
	return lg
}

func (s *XORMLogger) printSql(v ...interface{}) {
	var sql, params interface{}
	if len(v) > 0 {
		sql = v[0]
	} else {
		sql = ""
	}
	if len(v) > 1 {
		params = v[1]
	} else {
		params = nil
	}
	args := []interface{}{"sql", "", 0, sql, params, MaxInt64}
	s.logger.Print(args...)
}

// Error implement core.ILogger
func (s *XORMLogger) Error(v ...interface{}) {
	if s.level <= log.LOG_ERR {
		s.printSql(v...)
	}
	return
}

// Errorf implement core.ILogger
func (s *XORMLogger) Errorf(format string, v ...interface{}) {
	if s.level <= log.LOG_ERR {
		s.printSql(v...)
	}
	return
}

// Debug implement core.ILogger
func (s *XORMLogger) Debug(v ...interface{}) {
	if s.level <= log.LOG_DEBUG {
		s.printSql(v...)
	}
	return
}

// Debugf implement core.ILogger
func (s *XORMLogger) Debugf(format string, v ...interface{}) {
	if s.level <= log.LOG_DEBUG {
		s.printSql(v...)
	}
	return
}

// Info implement core.ILogger
func (s *XORMLogger) Info(v ...interface{}) {
	if s.level <= log.LOG_INFO {
		s.printSql(v...)
	}
	return
}

// Infof implement core.ILogger
func (s *XORMLogger) Infof(format string, v ...interface{}) {
	if s.level <= log.LOG_INFO {
		s.printSql(v...)
	}
	return
}

// Warn implement core.ILogger
func (s *XORMLogger) Warn(v ...interface{}) {
	if s.level <= log.LOG_WARNING {
		s.printSql(v...)
	}
	return
}

// Warnf implement core.ILogger
func (s *XORMLogger) Warnf(format string, v ...interface{}) {
	if s.level <= log.LOG_WARNING {
		s.printSql(v...)
	}
	return
}

// Level implement core.ILogger
func (s *XORMLogger) Level() log.LogLevel {
	return s.level
}

// SetLevel implement core.ILogger
func (s *XORMLogger) SetLevel(l log.LogLevel) {
	s.level = l
	return
}

// ShowSQL implement core.ILogger
func (s *XORMLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		s.showSQL = true
		return
	}
	s.showSQL = show[0]
}

// IsShowSQL implement core.ILogger
func (s *XORMLogger) IsShowSQL() bool {
	return s.showSQL
}
