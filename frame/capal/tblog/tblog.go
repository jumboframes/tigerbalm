package tblog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	DefaultLog *TbLog

	ErrUnsupportedLogLevel = errors.New("unsupported log level")
)

type Level int

const (
	LevelNull  Level = 0
	LevelTrace Level = 1
	LevelDebug Level = 2
	LevelInfo  Level = 3
	LevelWarn  Level = 4
	LevelError Level = 5
	LevelFatal Level = 6

	traceS = "TRACE"
	debugS = "DEBUG"
	infoS  = "INFO"
	warnS  = "WARN"
	errorS = "ERROR"
	fatalS = "FATAL"
)

var (
	levelStrings = map[Level]string{
		LevelTrace: traceS,
		LevelDebug: debugS,
		LevelInfo:  infoS,
		LevelWarn:  warnS,
		LevelError: errorS,
		LevelFatal: fatalS,
	}

	levelInts = map[string]Level{
		traceS: LevelTrace,
		debugS: LevelDebug,
		infoS:  LevelInfo,
		warnS:  LevelWarn,
		errorS: LevelError,
		fatalS: LevelFatal,
	}
)

func ParseLevel(levelS string) (Level, error) {
	level, ok := levelInts[strings.ToUpper(levelS)]
	if !ok {
		return LevelNull, ErrUnsupportedLogLevel
	}
	return level, nil
}

type TbLog struct {
	logger *log.Logger
	level  Level
	mu     sync.RWMutex
}

func init() {
	DefaultLog = NewTbLog()
}

func WithLevel(level Level) *TbLog {
	DefaultLog.mu.Lock()
	defer DefaultLog.mu.Unlock()
	DefaultLog.level = level
	return DefaultLog
}

func WithOutput(out io.Writer) *TbLog {
	DefaultLog.logger.SetOutput(out)
	return DefaultLog
}

func WithFlags(flag int) *TbLog {
	DefaultLog.logger.SetFlags(flag)
	return DefaultLog
}

func SetOutput(out io.Writer) {
	DefaultLog.logger.SetOutput(out)
}

func SetLevel(level Level) {
	DefaultLog.mu.Lock()
	defer DefaultLog.mu.Unlock()
	DefaultLog.level = level
}

func SetFlags(flag int) {
	DefaultLog.logger.SetFlags(flag)
}

func Println(level Level, v ...interface{}) {
	prefix, _ := levelStrings[level]
	prefix = fmt.Sprintf("%-6s", prefix)
	DefaultLog.outputln(level, prefix, v...)
}

func Printf(level Level, format string, v ...interface{}) {
	prefix, _ := levelStrings[level]
	prefix = fmt.Sprintf("%-6s", prefix)
	DefaultLog.outputf(false, level, prefix, format, v...)
}

func Tracef(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", traceS)
	DefaultLog.outputf(true, LevelTrace, prefix, format, v...)
}

func Debugf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", debugS)
	DefaultLog.outputf(true, LevelDebug, prefix, format, v...)
}

func Infof(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", infoS)
	DefaultLog.outputf(true, LevelInfo, prefix, format, v...)
}

func Warnf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", warnS)
	DefaultLog.outputf(true, LevelWarn, prefix, format, v...)
}

func Error(v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", errorS)
	DefaultLog.outputln(LevelError, prefix, v...)
}

func Errorf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", errorS)
	DefaultLog.outputf(true, LevelError, prefix, format, v...)
}

func Fatalf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", fatalS)
	DefaultLog.outputf(true, LevelFatal, prefix, format, v...)
}

func NewTbLog() *TbLog {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return &TbLog{
		logger: logger,
		level:  LevelTrace,
	}
}

func (tblog *TbLog) WithLevel(level Level) *TbLog {
	tblog.mu.Lock()
	defer tblog.mu.Unlock()
	tblog.level = level
	return tblog
}

func (tblog *TbLog) WithOutput(out io.Writer) *TbLog {
	tblog.logger.SetOutput(out)
	return tblog
}

func (tblog *TbLog) WithFlags(flag int) *TbLog {
	tblog.logger.SetFlags(flag)
	return tblog
}

func (tblog *TbLog) SetOutput(out io.Writer) {
	tblog.logger.SetOutput(out)
	return
}

func (tblog *TbLog) SetLevel(level Level) {
	tblog.mu.Lock()
	defer tblog.mu.Unlock()
	tblog.level = level
	return
}

func (tblog *TbLog) SetFlags(flag int) {
	tblog.logger.SetFlags(flag)
	return
}

func (tblog *TbLog) Println(level Level, v ...interface{}) {
	prefix, _ := levelStrings[level]
	prefix = fmt.Sprintf("%-6s", prefix)
	tblog.outputln(level, prefix, v...)
}

func (tblog *TbLog) Printf(level Level, format string, v ...interface{}) {
	prefix, _ := levelStrings[level]
	prefix = fmt.Sprintf("%-6s", prefix)
	tblog.outputf(false, level, prefix, format, v...)
}

func (tblog *TbLog) Tracef(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", traceS)
	tblog.outputf(true, LevelTrace, prefix, format, v...)
}

func (tblog *TbLog) Debugf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", debugS)
	tblog.outputf(true, LevelDebug, prefix, format, v...)
}

func (tblog *TbLog) Infof(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", infoS)
	tblog.outputf(true, LevelInfo, prefix, format, v...)
}

func (tblog *TbLog) Warnf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", warnS)
	tblog.outputf(true, LevelWarn, prefix, format, v...)
}

func (tblog *TbLog) Error(v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", errorS)
	tblog.outputln(LevelError, prefix, v...)
}

func (tblog *TbLog) Errorf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", errorS)
	tblog.outputf(true, LevelError, prefix, format, v...)
}

func (tblog *TbLog) Fatalf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("%-6s", fatalS)
	tblog.outputf(true, LevelFatal, prefix, format, v...)
}

func (tblog *TbLog) outputln(level Level, prefix string, v ...interface{}) {
	tblog.mu.RLock()
	defer tblog.mu.RUnlock()
	if level < tblog.level {
		return
	}

	line := fmt.Sprintln(v...)
	line = prefix + line
	tblog.logger.Output(2, line)
	if level == LevelFatal {
		os.Exit(1)
	}
}

func (tblog *TbLog) outputf(newline bool, level Level, prefix, format string, v ...interface{}) {
	tblog.mu.RLock()
	defer tblog.mu.RUnlock()
	if level < tblog.level {
		return
	}

	line := fmt.Sprintf(format, v...)
	line = prefix + line
	if newline {
		line += "\n"
	}
	tblog.logger.Output(2, line)
	if level == LevelFatal {
		os.Exit(1)
	}
}
