package logging

import (
	"fmt"
	"io"
	"time"
)

type Record struct {
	Time       time.Time
	Level      logLevel
	Message    string
	LoggerName string
}

type Emitter interface {
	Emit(string, *Record)
}

type Logger struct {
	Name     string
	Handlers map[string]Emitter
}

func NewLogger() *Logger {
	return &Logger{Handlers: make(map[string]Emitter)}
}

var DefaultLogger = NewLogger()

func (l *Logger) AddHandler(name string, h Emitter) {
	oldHandler, ok := l.Handlers[name]
	if ok {
		closer, ok := oldHandler.(io.Closer)
		if ok {
			_ = closer.Close()
		}
	}
	l.Handlers[name] = h
}

func (l *Logger) Log(level logLevel, format string, values ...interface{}) {
	rd := &Record{
		Time:       time.Now(),
		Level:      level,
		Message:    fmt.Sprintf(format, values...),
		LoggerName: l.Name,
	}
	for name, h := range l.Handlers {
		h.Emit(name, rd)
	}
}

func (l *Logger) Debug(format string, values ...interface{}) {
	l.Log(DEBUG, format, values...)
}

func (l *Logger) Info(format string, values ...interface{}) {
	l.Log(INFO, format, values...)
}

func (l *Logger) Warning(format string, values ...interface{}) {
	l.Log(WARNING, format, values...)
}

func (l *Logger) Error(format string, values ...interface{}) {
	l.Log(ERROR, format, values...)
}

func (l *Logger) ResetLogLevel(level string) {
	for _, e := range l.Handlers {
		if h, ok := e.(*Handler); ok {
			h.SetLevelString(level)
		}
	}
}

func AddHandler(name string, h Emitter) {
	DefaultLogger.AddHandler(name, h)
}

func Log(level logLevel, format string, values ...interface{}) {
	DefaultLogger.Log(level, format, values...)
}

func Debug(format string, values ...interface{}) {
	DefaultLogger.Log(DEBUG, format, values...)
}

func Info(format string, values ...interface{}) {
	DefaultLogger.Log(INFO, format, values...)
}

func Warning(format string, values ...interface{}) {
	DefaultLogger.Log(WARNING, format, values...)
}

func Error(format string, values ...interface{}) {
	DefaultLogger.Log(ERROR, format, values...)
}

func ResetLogLevel(level string) {
	DefaultLogger.ResetLogLevel(level)
}
