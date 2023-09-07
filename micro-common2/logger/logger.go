package logger

import (
	"context"
	"database/sql/driver"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

var MaxInt64 = ^int64(0)

var (
	version string
)

const (
	TraceIDKey = "trace_id"
	UserIDKey  = "user_id"
	TagKey     = "tag"
	VersionKey = "version"
	StackKey   = "stack"
)

type (
	traceIDKey struct{}
	userIDKey  struct{}
	tagKey     struct{}
	stackKey   struct{}
)

type Logger struct {
	logrus.Entry
	Level logrus.Level
}

func New() *Logger {
	timeFormat := "2006/01/02 15:04:05.000 -0700"
	if tmFormat, ok := viper.Get("logger.time_format").(string); ok && len(tmFormat) > 0 {
		timeFormat = strings.TrimSpace(tmFormat)
	}

	level := logrus.InfoLevel
	if lvStr, ok := viper.Get("logger.level").(string); ok {
		lvStr = strings.TrimSpace(strings.ToLower(lvStr))
		if lvStr == "warn" {
			level = logrus.WarnLevel
		} else if lvStr == "debug" {
			level = logrus.DebugLevel
		} else if lvStr == "error" {
			level = logrus.ErrorLevel
		} else if lvStr == "fatal" {
			level = logrus.FatalLevel
		} else if lvStr == "panic" {
			level = logrus.PanicLevel
		} else if lvStr == "trace" {
			level = logrus.TraceLevel
		}
	}
	var formatter logrus.Formatter
	if viper.GetString("logger.formatter") == "json" {
		formatter = &logrus.JSONFormatter{TimestampFormat: timeFormat}
	} else {
		formatter = &logrus.TextFormatter{TimestampFormat: timeFormat}
	}
	log := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}
	if grayAddr, ok := viper.Get("logger.graylog.addr").(string); ok && len(grayAddr) > 0 {
		grayHook := graylog.NewGraylogHook(grayAddr, nil)
		log.AddHook(grayHook)
	}
	lfMap := viper.GetStringMapString("logger.local.file.path")
	if nil != lfMap && len(lfMap) > 0 {
		viper.SetDefault("logger.local.file.rotation.hours", 24)
		viper.SetDefault("logger.local.file.rotation.count", 7)
		viper.SetDefault("logger.local.file.rotation.postfix", ".%Y%m%d%H%M")
		rotationHours := viper.GetInt("logger.local.file.rotation.hours")
		rotationCount := viper.GetInt("logger.local.file.rotation.count")
		rotationPostfix := viper.GetString("logger.local.file.rotation.postfix")
		writerMap := lfshook.WriterMap{}
		if v, ok := lfMap["panic"]; ok {
			writerMap[logrus.PanicLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),                                          // 为最新的日志建立软连接，以方便随着找到当前日志文件
				rotatelogs.WithRotationCount(uint(rotationCount)),                   // 设置文件清理前最多保存的个数，也可通过WithMaxAge设置最长保存时间，二者取其一
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour), // 设置日志分割的时间，例如一天一次
			)
		}
		if v, ok := lfMap["fatal"]; ok {
			writerMap[logrus.FatalLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),
				rotatelogs.WithRotationCount(uint(rotationCount)),
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
			)
		}
		if v, ok := lfMap["error"]; ok {
			writerMap[logrus.ErrorLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),
				rotatelogs.WithRotationCount(uint(rotationCount)),
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
			)
		}
		if v, ok := lfMap["warn"]; ok {
			writerMap[logrus.WarnLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),
				rotatelogs.WithRotationCount(uint(rotationCount)),
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
			)
		}
		if v, ok := lfMap["info"]; ok {
			writerMap[logrus.InfoLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),
				rotatelogs.WithRotationCount(uint(rotationCount)),
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
			)
		}
		if v, ok := lfMap["debug"]; ok {
			writerMap[logrus.DebugLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),
				rotatelogs.WithRotationCount(uint(rotationCount)),
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
			)
		}
		if v, ok := lfMap["trace"]; ok {
			writerMap[logrus.TraceLevel], _ = rotatelogs.New(
				v+rotationPostfix,
				rotatelogs.WithLinkName(v),
				rotatelogs.WithRotationCount(uint(rotationCount)),
				rotatelogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
			)
		}
		var lfFormatter logrus.Formatter
		if viper.GetString("logger.local.file.formatter") == "json" {
			lfFormatter = &logrus.JSONFormatter{TimestampFormat: timeFormat}
		} else {
			lfFormatter = &logrus.TextFormatter{TimestampFormat: timeFormat}
		}
		lfHook := lfshook.NewHook(writerMap, lfFormatter)
		log.AddHook(lfHook)
	}
	entry := logrus.NewEntry(log)

	extra := viper.GetStringMap("logger.extra")
	if nil != extra && len(extra) > 0 {
		entry = entry.WithFields(extra)
	}

	return &Logger{Entry: *entry, Level: level}
}

func (logger *Logger) Print(args ...interface{}) {
	if args == nil || len(args) == 0 {
		return
	}
	if tp, ok := args[0].(string); ok {
		tp = strings.ToLower(strings.TrimSpace(tp))
		if "sql" == tp && len(args) == 6 {
			logger.printSql(args...)
		} else {
			logger.SetCaller().Entry.Print(args...)
		}
	} else {
		logger.SetCaller().Entry.Print(args...)
	}
}

func (logger *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Entry: *logger.Entry.WithField(key, value)}
}

func (logger *Logger) WithFields(fields map[string]interface{}) *Logger {
	return &Logger{Entry: *logger.Entry.WithFields(fields)}
}

func (logger *Logger) WithError(err error) *Logger {
	return &Logger{Entry: *logger.Entry.WithError(err)}
}

func (logger *Logger) WithCaller(skip int) *Logger {
	if _, ok := logger.Data["codeline"]; ok {
		return logger
	}
	//for i := 0; i < 100; i++ {
	//	if _, file, line, ok := runtime.Caller(i); ok {
	//		if strings.Contains(file, "it.sz.cn") &&
	//			!strings.Contains(file, "pp/common-golang/logger") {
	//			return logger.
	//				WithField("codeline", fmt.Sprintf("%s:%d", file, line))
	//			//WithField("func", runtime.FuncForPC(pc).Name())
	//		}
	//	}
	//}
	if _, file, line, ok := runtime.Caller(skip); ok {
		return logger.
			WithField("codeline", fmt.Sprintf("%s:%d", file, line))
		//WithField("func", runtime.FuncForPC(pc).Name())
	}
	return logger
}

func (logger *Logger) SetCaller() *Logger {
	return logger.WithCaller(4)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Debugf(format, args...)
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.SetCaller().Entry.Infof(format, args...)
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Printf(format, args...)
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Warnf(format, args...)
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Warningf(format, args...)
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Errorf(format, args...)
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Fatalf(format, args...)
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.SetCaller().Entry.Panicf(format, args...)
}

func (logger *Logger) Debug(args ...interface{}) {
	logger.SetCaller().Entry.Debug(args...)
}

func (logger *Logger) Info(args ...interface{}) {
	logger.SetCaller().Entry.Info(args...)
}

func (logger *Logger) Warn(args ...interface{}) {
	logger.SetCaller().Entry.Warn(args...)
}

func (logger *Logger) Warning(args ...interface{}) {
	logger.SetCaller().Entry.Warning(args...)
}

func (logger *Logger) Error(args ...interface{}) {
	logger.SetCaller().Entry.Error(args...)
}

func (logger *Logger) Fatal(args ...interface{}) {
	logger.SetCaller().Entry.Fatal(args...)
}

func (logger *Logger) Panic(args ...interface{}) {
	logger.SetCaller().Entry.Panic(args...)
}

func (logger *Logger) Debugln(args ...interface{}) {
	logger.SetCaller().Entry.Debugln(args...)
}

func (logger *Logger) Infoln(args ...interface{}) {
	logger.SetCaller().Entry.Infoln(args...)
}

func (logger *Logger) Println(args ...interface{}) {
	logger.SetCaller().Entry.Println(args...)
}

func (logger *Logger) Warnln(args ...interface{}) {
	logger.SetCaller().Entry.Warnln(args...)
}

func (logger *Logger) Warningln(args ...interface{}) {
	logger.SetCaller().Entry.Warningln(args...)
}

func (logger *Logger) Errorln(args ...interface{}) {
	logger.SetCaller().Entry.Errorln(args...)
}

func (logger *Logger) Fatalln(args ...interface{}) {
	logger.SetCaller().Entry.Fatalln(args...)
}

func (logger *Logger) Panicln(args ...interface{}) {
	logger.SetCaller().Entry.Panicln(args...)
}

// NewTraceIDContext 创建跟踪ID上下文
func NewTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// FromTraceIDContext 从上下文中获取跟踪ID
func FromTraceIDContext(ctx context.Context) string {
	v := ctx.Value(traceIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewUserIDContext 创建用户ID上下文
func NewUserIDContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// FromUserIDContext 从上下文中获取用户ID
func FromUserIDContext(ctx context.Context) string {
	v := ctx.Value(userIDKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewTagContext 创建Tag上下文
func NewTagContext(ctx context.Context, tag string) context.Context {
	return context.WithValue(ctx, tagKey{}, tag)
}

// FromTagContext 从上下文中获取Tag
func FromTagContext(ctx context.Context) string {
	v := ctx.Value(tagKey{})
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewStackContext 创建Stack上下文
func NewStackContext(ctx context.Context, stack error) context.Context {
	return context.WithValue(ctx, stackKey{}, stack)
}

// FromStackContext 从上下文中获取Stack
func FromStackContext(ctx context.Context) error {
	v := ctx.Value(stackKey{})
	if v != nil {
		if s, ok := v.(error); ok {
			return s
		}
	}
	return nil
}

func (logger *Logger) WithTrace(ctx context.Context) *Logger {
	if ctx == nil {
		ctx = context.Background()
	}

	fields := map[string]interface{}{
		VersionKey: version,
	}

	if v := FromTraceIDContext(ctx); v != "" {
		fields[TraceIDKey] = v
	}

	if v := FromUserIDContext(ctx); v != "" {
		fields[UserIDKey] = v
	}

	if v := FromTagContext(ctx); v != "" {
		fields[TagKey] = v
	}

	if v := FromStackContext(ctx); v != nil {
		fields[StackKey] = fmt.Sprintf("%+v", v)
	}

	logger.Entry.WithContext(ctx).WithFields(fields)

	return logger

}

func (logger *Logger) V(v int) bool {
	return false
}

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

func (logger *Logger) printSql(args ...interface{}) {
	length := len(args)
	var (
		codeLine, sql string
		params        []interface{}
		latency       time.Duration
		rows          int64
		ok            bool
	)
	if length > 1 {
		codeLine, _ = args[1].(string)
	}
	if length > 2 {
		latency, _ = args[2].(time.Duration)
	}
	if length > 3 {
		sql, ok = args[3].(string)
		if ok {
			sql = strings.TrimSpace(strings.Replace(strings.Replace(strings.Replace(sql, "\r\n", " ", -1), "\n", " ", -1), "\t", " ", -1))
		}
	}
	if length > 4 {
		params, _ = args[4].([]interface{})
	}
	if length > 5 {
		rows, _ = args[5].(int64)
	}
	lg := logger.
		WithField("tag", "SQL").
		WithField("sql", logger.getSql(sql, params))
	if len(codeLine) > 0 {
		lg = lg.WithField("codeline", strings.TrimSpace(codeLine))
	} else {
		lg = lg.WithCaller(9)
	}
	if latency > 0 {
		lg = lg.WithField("latency", fmt.Sprintf("%v", latency))
	}
	if rows != MaxInt64 {
		lg = lg.WithField("rows", fmt.Sprintf("%d rows affected or returned", rows))
	}
	if len(params) <= 0 {
		lg.Info(fmt.Sprintf("%s;", sql))
	} else {
		lg.Info(fmt.Sprintf("%s %v", sql, params))
	}
}

func (logger *Logger) getSql(originSql string, params []interface{}) string {
	var formattedValues []string
	for _, value := range params {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		if indirectValue.IsValid() {
			value = indirectValue.Interface()
			if t, ok := value.(time.Time); ok {
				formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
			} else if b, ok := value.([]byte); ok {
				if str := string(b); logger.isPrintable(str) {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
				} else {
					formattedValues = append(formattedValues, "'<binary>'")
				}
			} else if r, ok := value.(driver.Valuer); ok {
				if value, err := r.Value(); err == nil && value != nil {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			} else {
				formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
			}
		} else {
			formattedValues = append(formattedValues, "NULL")
		}
	}
	if nil == formattedValues {
		return ""
	}

	var sql string
	// differentiate between $n placeholders or else treat like ?
	if numericPlaceHolderRegexp.MatchString(originSql) {
		for index, value := range formattedValues {
			placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
			sql = regexp.MustCompile(placeholder).ReplaceAllString(originSql, value+"$1")
		}
	} else {
		formattedValuesLength := len(formattedValues)
		for index, value := range sqlRegexp.Split(originSql, -1) {
			sql += value
			if index < formattedValuesLength {
				sql += formattedValues[index]
			}
		}
	}
	return sql
}

func (logger *Logger) isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
