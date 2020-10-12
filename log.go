/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package gd

import (
	"encoding/xml"
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/utls"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	l             = sync.Mutex{}
	logConfigFile = "conf/log.xml"
	defaultLogDir = "log"
	defaultFormat = "%L	%D %T	%l	%I	%G	%M	%S"
)

type xmlLoggerConfig struct {
	ScribeCategory string      `xml:"scribeCategory"`
	Filter         []xmlFilter `xml:"filter"`
}

type xmlProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlFilter struct {
	Enabled  string        `xml:"enabled,attr"`
	Tag      string        `xml:"tag"`
	Level    string        `xml:"level"`
	Type     string        `xml:"type"`
	Property []xmlProperty `xml:"property"`
}

func getInfoFileName(binName string, port int) string {
	if port == 0 {
		return fmt.Sprintf("%s.log", binName)
	}
	return fmt.Sprintf("%s_%d.log", binName, port)
}

func getWarnFileName(binName string, port int) string {
	if port == 0 {
		return fmt.Sprintf("%s_err.log", binName)
	}
	return fmt.Sprintf("%s_err_%d.log", binName, port)
}

func restoreLogConfig(configFilePath string, binName string, port int, logLevel string, logDir string) error {
	l.Lock()
	defer l.Unlock()
	if logDir == "" {
		logDir = defaultLogDir
	}

	if configFilePath == "" {
		configFilePath = logConfigFile
	}

	if logLevel == "" {
		logLevel = "DEBUG"
	}

	if binName == "" {
		ex, err := os.Executable()
		if err != nil {
			return err
		}
		exPath := filepath.Dir(ex)
		if strings.Contains(exPath, "/") {
			ex = ex[len(exPath)+1:]
		}
		binName = ex
	}

	if logLevel != "DEBUG" && logLevel != "INFO" && logLevel != "WARNING" && logLevel != "ERROR" {
		return fmt.Errorf("invalid log level %v", logLevel)
	}

	infoFileName := getInfoFileName(binName, port)
	warnFileName := getWarnFileName(binName, port)

	var filters []xmlFilter
	// stdout
	stdout := xmlFilter{
		Enabled: "false",
		Tag:     "stdout",
		Level:   "INFO",
		Type:    "console",
	}
	filters = append(filters, stdout)
	// info
	info := xmlFilter{
		Enabled: "true",
		Tag:     "service",
		Level:   logLevel,
		Type:    "file",
		Property: []xmlProperty{
			xmlProperty{Name: "filename", Value: fmt.Sprintf("%s/%s", logDir, infoFileName)},
			xmlProperty{Name: "format", Value: defaultFormat},
			xmlProperty{Name: "rotate", Value: "true"},
			xmlProperty{Name: "maxsize", Value: "0M"},
			xmlProperty{Name: "maxlines", Value: "0K"},
			xmlProperty{Name: "hourly", Value: "true"},
		},
	}
	filters = append(filters, info)
	// warn
	warn := xmlFilter{
		Enabled: "true",
		Tag:     "service_err",
		Level:   "WARNING",
		Type:    "file",
		Property: []xmlProperty{
			xmlProperty{Name: "filename", Value: fmt.Sprintf("%s/%s", logDir, warnFileName)},
			xmlProperty{Name: "format", Value: defaultFormat},
			xmlProperty{Name: "rotate", Value: "true"},
			xmlProperty{Name: "maxsize", Value: "0M"},
			xmlProperty{Name: "maxlines", Value: "0K"},
			xmlProperty{Name: "hourly", Value: "true"},
		},
	}
	filters = append(filters, warn)

	c := &xmlLoggerConfig{
		Filter: filters,
	}

	bts, err := xml.Marshal(c)
	if err != nil {
		return err
	}

	err = utls.Store2File(configFilePath, string(bts))
	if err != nil {
		return err
	}

	return nil
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfiguration(filename string) {
	dlog.LoadConfiguration(filename)
}

// wrap log debug
func Debug(arg0 interface{}, args ...interface{}) {
	dlog.Debug(arg0, args)
}

func Crash(args ...interface{}) {
	dlog.Crash(args)
}

// Logs the given message and crashes the program
func Crashf(format string, args ...interface{}) {
	dlog.Crashf(format, args)
}

// Compatibility with `log`
func Exit(args ...interface{}) {
	dlog.Exit(args)
}

// Compatibility with `log`
func Exitf(format string, args ...interface{}) {
	dlog.Exitf(format, args)
}

// Compatibility with `log`
func Stderr(args ...interface{}) {
	dlog.Stderr(args)
}

// Compatibility with `log`
func Stderrf(format string, args ...interface{}) {
	dlog.Stderrf(format, args)
}

// Compatibility with `log`
func Stdout(args ...interface{}) {
	dlog.Stdout(args)
}

// Compatibility with `log`
func Stdoutf(format string, args ...interface{}) {
	dlog.Stdoutf(format, args)
}

func GetLevel() string {
	return dlog.GetLevel()
}

func SetLevel(lvl int) {
	dlog.SetLevel(lvl)
}

// Send a log message manually
// Wrapper for (*Logger).Log
func Log(lvl dlog.Level, source, message string) {
	dlog.Log(lvl, source, message)
}

// Send a formatted log message easily
// Wrapper for (*Logger).Logf
func Logf(lvl dlog.Level, format string, args ...interface{}) {
	dlog.Logf(lvl, format, args)
}

// Send a closure log message
// Wrapper for (*Logger).Logc
func Logc(lvl dlog.Level, closure func() string) {
	dlog.Logc(lvl, closure)
}

// Utility for finest log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	dlog.Finest(arg0, args)
}

// Utility for fine log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	dlog.Fine(arg0, args)
}

func DebugT(tag string, arg0 interface{}, args ...interface{}) {
	dlog.DebugT(tag, arg0, args)
}

// Utility for trace log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	dlog.Trace(arg0, args)
}

func TraceT(tag string, arg0 interface{}, args ...interface{}) {
	dlog.TraceT(tag, arg0, args)
}

// Utility for info log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	dlog.Info(arg0, args)
}

func InfoT(tag string, arg0 interface{}, args ...interface{}) {
	dlog.InfoT(tag, arg0, args)
}

// Utility for warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) {
	dlog.Warn(arg0, args)
}

func WarnT(tag string, arg0 interface{}, args ...interface{}) {
	dlog.WarnT(tag, arg0, args)
}

// Utility for error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) {
	dlog.Error(arg0, args)
}

func ErrorT(tag string, arg0 interface{}, args ...interface{}) {
	dlog.ErrorT(tag, arg0, args)
}

// Utility for critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) {
	dlog.Critical(arg0, args)
}

func CriticalT(tag string, arg0 interface{}, args ...interface{}) {
	dlog.CriticalT(tag, arg0, args)
}
