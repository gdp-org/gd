package logging

import (
	"testing"
)

func TestLevelToString(t *testing.T) {
	data := map[logLevel]string{
		DEBUG:         "DEBUG",
		INFO:          "INFO ",
		WARNING:       "WARN ",
		ERROR:         "ERROR",
		logLevel(250): "DISABLE",
	}
	for level, str := range data {
		if level.String() != str {
			t.Error(level.String())
		}
	}
}

func TestStringToLevel(t *testing.T) {
	data := map[string]logLevel{
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARN":    WARNING,
		"WARNING": WARNING,
		"ERROR":   ERROR,
		"":        DISABLE,
	}
	for str, level := range data {
		if stringToLogLevel(str) != level {
			t.Error(str)
		}
	}
}

func TestLevelRange(t *testing.T) {
	lr := levelRange{INFO, WARNING}
	if lr.contains(ERROR) {
		t.Error("TestLevelRange Fail")
	}
	if lr.contains(DEBUG) {
		t.Error("TestLevelRange Fail")
	}
	if !lr.contains(INFO) {
		t.Error("TestLevelRange Fail")
	}
}
