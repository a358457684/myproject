package log

import (
	"common/config"
	"context"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/suiyunonghen/DxCommonLib"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//控制台Hook，目的是将错误和非错误日志分开
type consoleHook struct {
	formatter logrus.Formatter
}

func (chook *consoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (chook *consoleHook) Fire(entry *logrus.Entry) error {
	bt, err := chook.formatter.Format(entry)
	if err != nil {
		return err
	}
	if entry.Level < logrus.WarnLevel {
		_, err = os.Stderr.Write(bt)
	} else {
		_, err = os.Stdout.Write(bt)
	}
	return err
}

type Logger struct {
	showcaller bool
	rootentry  *logrus.Entry
	l          logrus.Logger
}

//空的格式化处理，目的是让Hook了之后，不走主日志模块
type emptyFormater struct {
}

func (formater emptyFormater) Format(*logrus.Entry) ([]byte, error) {
	return nil, nil
}

type lazyWriter struct {
	logWriter []io.Writer
	wChan     chan []byte
}

func (w *lazyWriter) Write(p []byte) (n int, err error) {
	select {
	case w.wChan <- p:
		return len(p), nil
	case <-DxCommonLib.After(time.Second):
		return 0, nil
	}
}

func (w *lazyWriter) lazyWrite() {
	buf := make([]byte, 0, 10240)
	ltime := time.Now()
	for {
		select {
		case rp, ok := <-w.wChan:
			if !ok {
				return
			}
			buf = append(buf, rp...)
			ntime := time.Now()
			if ntime.Sub(ltime) >= time.Second*5 || len(buf) >= 10240 {
				for _, writer := range w.logWriter {
					writer.Write(buf)
				}
				buf = buf[:0]
				ltime = ntime
			}
		case <-DxCommonLib.After(time.Second * 6):
			ntime := time.Now()
			if ntime.Sub(ltime) >= time.Second*5 {
				for _, writer := range w.logWriter {
					writer.Write(buf)
				}
				buf = buf[:0]
				ltime = ntime
			}
		}
	}
}

func NewLog(logcfg *config.LogOptions) *Logger {
	result := Logger{}
	result.showcaller = logcfg.ShowCaller
	var formatter logrus.Formatter
	if logcfg.JsonFormat {
		formatter = &logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.000"}
	} else {
		formatter = &logrus.TextFormatter{DisableColors: !logcfg.ColorLog, DisableQuote: true, FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.000"}
	}
	result.l.Formatter = &emptyFormater{} //指定为空的format,目的是不执行主处理的格式化以及输出
	createLogFileWriter := func() io.Writer {
		if logcfg.SplitTime == 0 {
			if logcfg.SplitSize == 0 {
				logcfg.SplitSize = 1024 * 10000
			}
			return New(logcfg.SplitSize, logcfg.File)
		}
		fileNameinfos := strings.FieldsFunc(logcfg.File, func(r rune) bool {
			return r == '.'
		})
		fl := len(fileNameinfos)
		if fl > 1 { //有扩展名
			fileNameinfos[fl-2] = fileNameinfos[fl-2] + "%Y%m%d%H"
			fileNameinfos[fl-1] = "." + fileNameinfos[fl-1]
		}
		newfileName := strings.Join(fileNameinfos, "")
		if logcfg.File[0] == '.' {
			newfileName = "." + newfileName
		}
		w, err := rotatelogs.New(newfileName, rotatelogs.WithRotationTime(time.Hour*time.Duration(logcfg.SplitTime)), rotatelogs.WithMaxAge(time.Hour*24*7))
		if err != nil {
			fmt.Println("创建日志文件失败：", err.Error())
			return nil
		}
		return w
	}

	if logcfg.LazyWrite {
		//懒写入模式
		lwriter := &lazyWriter{
			logWriter: make([]io.Writer, 0, 4),
			wChan:     make(chan []byte, 32),
		}
		if logcfg.ShowConsole {
			lwriter.logWriter = append(lwriter.logWriter, os.Stderr)
		}
		if logcfg.File != "" {
			if filewriter := createLogFileWriter(); filewriter != nil {
				lwriter.logWriter = append(lwriter.logWriter, filewriter)
			}
		}
		go lwriter.lazyWrite()
		result.l.Out = lwriter
	} else {
		result.l.Out = os.Stderr
	}

	result.l.Hooks = make(logrus.LevelHooks)
	result.l.ExitFunc = os.Exit
	LogLevel, err := logrus.ParseLevel(logcfg.Level)
	if err != nil {
		LogLevel = logrus.InfoLevel
	}
	result.l.Level = LogLevel

	fields := make(logrus.Fields, 3)
	if logcfg.Project != "" {
		fields["project"] = logcfg.Project
	}
	if logcfg.Author != "" {
		fields["author"] = logcfg.Author
	}
	if logcfg.Machine != "" {
		fields["machine"] = logcfg.Machine
	}
	if len(fields) > 0 {
		result.rootentry = result.l.WithFields(fields)
	} else {
		result.rootentry = nil
	}

	if !logcfg.LazyWrite && logcfg.File != "" {
		filewriter := createLogFileWriter()
		if filewriter != nil {
			lfsHook := lfshook.NewHook(lfshook.WriterMap{
				logrus.DebugLevel: filewriter,
				logrus.InfoLevel:  filewriter,
				logrus.WarnLevel:  filewriter,
				logrus.ErrorLevel: filewriter,
				logrus.FatalLevel: filewriter,
				logrus.PanicLevel: filewriter,
			}, formatter)
			result.l.AddHook(lfsHook)
		}
	}
	if logcfg.ShowConsole {
		//将错误和其他信息分开来显示
		result.l.AddHook(&consoleHook{formatter})
	}
	return &result
}

var bufferpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 256)
	},
}

func (l *Logger) caller(offset int) (string, string) {
	//找到他的上上上一级的调用位置,0是当前位置,1是上一个位置是withfield，或者log输出，所以是上上一级
	pc, file, line, ok := runtime.Caller(2 + offset)
	if !ok {
		return "", ""
	}
	buffer := bufferpool.Get().([]byte)
	idx := strings.LastIndexByte(file, '/')
	if idx == -1 {
		buffer = append(buffer[:0], file...)
	} else {
		idx = strings.LastIndexByte(file[:idx], '/')
		if idx == -1 {
			buffer = append(buffer[:0], file...)
		} else {
			buffer = append(buffer[:0], file[idx+1:]...)
		}
	}
	funcName := runtime.FuncForPC(pc).Name()
	buffer = append(buffer, ':')
	buffer = strconv.AppendInt(buffer, int64(line), 10)
	result := string(buffer)
	bufferpool.Put(buffer)
	idx = strings.IndexByte(funcName, '.')
	if idx > 0 {
		funcName = funcName[idx+1:]
	}
	return result, funcName
}

func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.withField(1, key, value)
}

func (l *Logger) withField(offset int, key string, value interface{}) *logrus.Entry {
	if l.showcaller {
		fields := make(logrus.Fields, 2)
		fields[key] = value
		fields["caller"], fields["func"] = l.caller(offset)
		if l.rootentry != nil {
			return l.rootentry.WithFields(fields)
		}
		return l.l.WithFields(fields)
	}

	if l.rootentry != nil {
		return l.rootentry.WithField(key, value)
	}
	return l.l.WithField(key, value)
}

func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.withFields(1, fields)
}

func (l *Logger) withFields(offset int, fields logrus.Fields) *logrus.Entry {
	if l.showcaller {
		fields["caller"], fields["func"] = l.caller(offset)
	}
	if l.rootentry != nil {
		return l.rootentry.WithFields(fields)
	}
	return l.l.WithFields(fields)
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.withError(1, err)
}

func (l *Logger) withError(offset int, err error) *logrus.Entry {
	if l.showcaller {
		fields := make(logrus.Fields, 2)
		fields[logrus.ErrorKey] = err
		fields["caller"], fields["func"] = l.caller(offset)
		if l.rootentry != nil {
			return l.rootentry.WithFields(fields)
		}
		return l.l.WithFields(fields)
	}

	if l.rootentry != nil {
		return l.rootentry.WithField(logrus.ErrorKey, err)
	}
	return l.l.WithField(logrus.ErrorKey, err)
}

func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	return l.withContext(1, ctx)
}

func (l *Logger) withContext(offset int, ctx context.Context) *logrus.Entry {
	newctx := ctx
	if l.showcaller {
		caller, funcname := l.caller(offset)
		newctx = context.WithValue(ctx, "caller", caller)
		newctx = context.WithValue(newctx, "func", funcname)
	}
	if l.rootentry != nil {
		return l.rootentry.WithContext(newctx)
	}
	return l.l.WithContext(newctx)
}

func (l *Logger) currentCallerEntry(offset int) *logrus.Entry {
	fields := make(logrus.Fields, 2)
	fields["caller"], fields["func"] = l.caller(offset)
	if l.rootentry != nil {
		return l.rootentry.WithFields(fields)
	}
	return l.l.WithFields(fields)
}

func (l *Logger) logf(calloffset int, level logrus.Level, format string, args ...interface{}) {
	if l.showcaller {
		l.currentCallerEntry(calloffset).Logf(level, format, args...)
	} else if l.rootentry != nil {
		l.rootentry.Logf(level, format, args...)
	} else {
		l.l.Logf(level, format, args...)
	}
}

func (l *Logger) Logf(level logrus.Level, format string, args ...interface{}) {
	l.logf(2, level, format, args...)
}

func (logger *Logger) Tracef(format string, args ...interface{}) {
	logger.logf(2, logrus.TraceLevel, format, args...)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.logf(2, logrus.DebugLevel, format, args...)
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.logf(2, logrus.InfoLevel, format, args...)
}

func (logger *Logger) Printf(format string, args ...interface{}) {
	logger.logf(2, logrus.InfoLevel, format, args...)
}

func (logger *Logger) Warnf(format string, args ...interface{}) {
	logger.logf(2, logrus.WarnLevel, format, args...)
}

func (logger *Logger) Warningf(format string, args ...interface{}) {
	logger.logf(2, logrus.WarnLevel, format, args...)
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.logf(2, logrus.ErrorLevel, format, args...)
}

func (logger *Logger) Fatalf(format string, args ...interface{}) {
	logger.logf(2, logrus.FatalLevel, format, args...)
	logger.l.Exit(1)
}

func (logger *Logger) Panicf(format string, args ...interface{}) {
	logger.logf(2, logrus.PanicLevel, format, args...)
}

func (logger *Logger) log(offset int, level logrus.Level, args ...interface{}) {
	if logger.showcaller {
		logger.currentCallerEntry(offset).Log(level, args...)
	} else if logger.rootentry != nil {
		logger.rootentry.Log(level, args...)
	} else {
		logger.l.Log(level, args...)
	}
}

func (logger *Logger) Log(level logrus.Level, args ...interface{}) {
	logger.log(2, level, args...)
}

func (logger *Logger) Trace(args ...interface{}) {
	logger.log(2, logrus.TraceLevel, args...)
}

func (logger *Logger) Debug(args ...interface{}) {
	logger.log(2, logrus.DebugLevel, args...)
}

func (logger *Logger) Info(args ...interface{}) {
	logger.log(2, logrus.InfoLevel, args...)
}

func (logger *Logger) Print(args ...interface{}) {
	logger.log(2, logrus.InfoLevel, args...)
}

func (logger *Logger) Warn(args ...interface{}) {
	logger.log(2, logrus.WarnLevel, args...)
}

func (logger *Logger) Warning(args ...interface{}) {
	logger.log(2, logrus.WarnLevel, args...)
}

func (logger *Logger) Error(args ...interface{}) {
	logger.log(2, logrus.ErrorLevel, args...)
}

func (logger *Logger) Fatal(args ...interface{}) {
	logger.log(2, logrus.FatalLevel, args...)
	logger.l.Exit(1)
}

func (logger *Logger) Panic(args ...interface{}) {
	logger.log(2, logrus.PanicLevel, args...)
}

func (logger *Logger) logln(offset int, level logrus.Level, args ...interface{}) {
	if logger.showcaller {
		logger.currentCallerEntry(offset).Logln(level, args...)
	} else if logger.rootentry != nil {
		logger.rootentry.Logln(level, args...)
	} else {
		logger.l.Logln(level, args...)
	}
}

func (logger *Logger) Logln(level logrus.Level, args ...interface{}) {
	logger.logln(2, level, args...)
}

func (logger *Logger) Traceln(args ...interface{}) {
	logger.logln(2, logrus.TraceLevel, args...)
}

func (logger *Logger) Debugln(args ...interface{}) {
	logger.logln(2, logrus.DebugLevel, args...)
}

func (logger *Logger) Infoln(args ...interface{}) {
	logger.logln(2, logrus.InfoLevel, args...)
}

func (logger *Logger) Println(args ...interface{}) {
	logger.logln(2, logrus.InfoLevel, args...)
}

func (logger *Logger) Warnln(args ...interface{}) {
	logger.logln(2, logrus.WarnLevel, args...)
}

func (logger *Logger) Warningln(args ...interface{}) {
	logger.logln(2, logrus.WarnLevel, args...)
}

func (logger *Logger) Errorln(args ...interface{}) {
	logger.logln(2, logrus.ErrorLevel, args...)
}

func (logger *Logger) Fatalln(args ...interface{}) {
	logger.logln(2, logrus.FatalLevel, args...)
	logger.l.Exit(1)
}

func (logger *Logger) Panicln(args ...interface{}) {
	logger.logln(2, logrus.PanicLevel, args...)
}
