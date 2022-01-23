package tblog

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Level int

const (
	Trace Level = 1
	Debug Level = 2
	Info  Level = 3
	Warn  Level = 4
	Error Level = 5
	Fatal Level = 6
)

var (
	levelMaps = map[Level]string{
		Trace: "TRACE",
		Debug: "DEBUG",
		Info:  "INFO",
		Warn:  "WARN",
		Error: "ERROR",
		Fatal: "FATAL",
	}
)

type Tblog struct {
	logger *log.Logger
	level  Level
	mu     sync.RWMutex
}

func NewTblog() *Tblog {
	//logger := log.New(os.Stdout, fmt.Sprintf("%6s", "TRACE"), log.LstdFlags)
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return &Tblog{
		logger: logger,
		level:  Trace,
	}
}

func (tblog *Tblog) WithLevel(level Level) *Tblog {
	tblog.mu.Lock()
	defer tblog.mu.Unlock()
	tblog.level = level
	return tblog
}

func (tblog *Tblog) WithOutput(out io.Writer) *Tblog {
	tblog.logger.SetOutput(out)
	return tblog
}

func (tblog *Tblog) WithFlags(flag int) *Tblog {
	tblog.logger.SetFlags(flag)
	return tblog
}

func (tblog *Tblog) SetOutput(out io.Writer) {
	tblog.logger.SetOutput(out)
	return
}

func (tblog *Tblog) SetLevel(level Level) {
	tblog.mu.Lock()
	defer tblog.mu.Unlock()
	tblog.level = level
	return
}

func (tblog *Tblog) SetFlags(flag int) {
	tblog.logger.SetFlags(flag)
	return
}

func (tblog *Tblog) Println(level Level, v ...interface{}) {
	prefix, _ := levelMaps[level]
	prefix = fmt.Sprintf("%-6s", prefix)
	tblog.mu.RLock()
	defer tblog.mu.RUnlock()
	if level < tblog.level {
		return
	}
	line := fmt.Sprintln(v...)
	line = prefix + line

	if level == Fatal {
		tblog.logger.Output(2, line)
		os.Exit(1)
	} else {
		tblog.logger.Output(2, line)
	}
}

func (tblog *Tblog) Printf(level Level, format string, v ...interface{}) {
	prefix, _ := levelMaps[level]
	prefix = fmt.Sprintf("%-6s", prefix)
	tblog.mu.RLock()
	defer tblog.mu.RUnlock()
	if level < tblog.level {
		return
	}
	line := fmt.Sprintf(format, v...)
	line = prefix + line
	if level == Fatal {
		tblog.logger.Output(2, line)
		os.Exit(1)
	} else {
		tblog.logger.Output(2, line)
	}
}
