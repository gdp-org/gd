package logging

import (
	"strings"
)

type logLevel uint8

const (
	DEBUG   logLevel = 1
	INFO    logLevel = 2
	WARNING logLevel = 3
	ERROR   logLevel = 4
	DISABLE logLevel = 255
)

func StringToLogLevel(s string) logLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return DISABLE
	}
}

func (level *logLevel) String() string {
	switch *level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARNING:
		return "WARN "
	case ERROR:
		return "ERROR"
	default:
		return "DISABLE"
	}
}

type levelRange struct {
	minLevel logLevel
	maxLevel logLevel
}

func (lr *levelRange) contains(level logLevel) bool {
	return level >= lr.minLevel && level <= lr.maxLevel
}
