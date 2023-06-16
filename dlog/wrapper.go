/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package dlog

import (
	"fmt"
	"github.com/chuck1024/gd/runtime/gl"
	"os"
	"strings"
)

var (
	Global Logger
)

func init() {
	Global = NewDefaultLogger(DEBUG)
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfiguration(filename string) {
	Global.LoadConfiguration(filename)
}

func LoadConfigurationByXml(xc *XmlLoggerConfig) {
	Global.LoadConfigurationByXml(xc)
}

// Wrapper for (*Logger).AddFilter
func AddFilter(name string, lvl Level, writer LogWriter) {
	Global.AddFilter(name, lvl, writer)
}

// Wrapper for (*Logger).Close (closes and removes all logwriters)
func Close() {
	Global.Close()
}

func Crash(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(CRITICAL, strings.Repeat(" %v", len(args))[1:], args...)
		panic(fmt.Sprintf(strings.Repeat(" %v", len(args))[1:], args...))
	} else {
		panic(args)
	}
}

func IsEnabledFor(lvl Level) bool {
	return Global.IsEnabledFor(lvl)
}

// Logs the given message and crashes the program
func Crashf(format string, args ...interface{}) {
	Global.intLogf(CRITICAL, format, args...)
	Global.Close() // so that hopefully the messages get logged
	panic(fmt.Sprintf(format, args...))
}

// Compatibility with `log`
func Exit(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args...)
	}
	Global.Close() // so that hopefully the messages get logged
	os.Exit(0)
}

// Compatibility with `log`
func Exitf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args...)
	Global.Close() // so that hopefully the messages get logged
	os.Exit(0)
}

// Compatibility with `log`
func Stderr(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args...)
	}
}

// Compatibility with `log`
func Stderrf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args...)
}

// Compatibility with `log`
func Stdout(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(INFO, strings.Repeat(" %v", len(args))[1:], args...)
	}
}

// Compatibility with `log`
func Stdoutf(format string, args ...interface{}) {
	Global.intLogf(INFO, format, args...)
}

func GetLevel() string {
	var ret string
	for tag, filter := range Global {
		ret = ret + tag + ":" + filter.Level.String() + ","
	}
	return ret
}

func SetLevel(lvl int) {
	level := Level(lvl)
	for _, filter := range Global {
		filter.Level = level
	}
}

// Send a log message manually
// Wrapper for (*Logger).Log
func Log(lvl Level, source, message string) {
	Global.Log(lvl, source, message)
}

// Send a formatted log message easily
// Wrapper for (*Logger).Logf
func Logf(lvl Level, format string, args ...interface{}) {
	Global.intLogf(lvl, format, args...)
}

// Send a closure log message
// Wrapper for (*Logger).Logc
func Logc(lvl Level, closure func() string) {
	Global.intLogc(lvl, closure)
}

// Utility for finest log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINEST
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for fine log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for debug log messages
// When given a string as the first argument, this behaves like Logf but with the DEBUG log level (e.g. the first argument is interpreted as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned by the closure iff it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments formatted with %v and separated by spaces (ala Sprint).
// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	tag := ""
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogcTag(tag, clientIp, logId, lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func DebugT(tag string, arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogcTag(tag, clientIp, logId, lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for trace log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	const (
		lvl = TRACE
	)
	tag := ""
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogcTag(tag, clientIp, logId, lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func TraceT(tag string, arg0 interface{}, args ...interface{}) {
	const (
		lvl = TRACE
	)
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogcTag(tag, clientIp, logId, lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for info log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	tag := ""
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogcTag(tag, clientIp, logId, lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func InfoT(tag string, arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogcTag(tag, clientIp, logId, lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) {
	const (
		lvl = WARNING
	)
	tag := ""
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogfTag(tag, clientIp, logId, lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

func WarnT(tag string, arg0 interface{}, args ...interface{}) {
	const (
		lvl = WARNING
	)
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogfTag(tag, clientIp, logId, lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) {
	const (
		lvl = ERROR
	)
	tag := ""
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogfTag(tag, clientIp, logId, lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

func ErrorT(tag string, arg0 interface{}, args ...interface{}) {
	const (
		lvl = ERROR
	)
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogfTag(tag, clientIp, logId, lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) {
	const (
		lvl = CRITICAL
	)
	tag := ""
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogfTag(tag, clientIp, logId, lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

func CriticalT(tag string, arg0 interface{}, args ...interface{}) {
	const (
		lvl = CRITICAL
	)
	clientIp, logId := batchGetGl()
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogfTag(tag, clientIp, logId, lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogfTag(tag, clientIp, logId, lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogfTag(tag, clientIp, logId, lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}

}

var batchGLKeys = []interface{}{
	gl.ClientIp,
	gl.LogId,
}

func batchGetGl() (ip, logId string) {
	vs := gl.BatchGet(batchGLKeys)
	if vs == nil {
		return
	}
	ipo, ok := vs[gl.ClientIp]
	if ok && ipo != nil {
		ip, _ = ipo.(string)
	}
	logIdo, ok := vs[gl.LogId]
	if ok && logIdo != nil {
		logId, _ = logIdo.(string)
	}
	return
}
