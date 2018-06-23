package logging

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

var h *Handler
var b *bytes.Buffer

func init() {
	b = bytes.NewBuffer(nil)
	h = NewHandler(b)
	DisableStdout()
	AddHandler("", h)
}

func TestSetLevel(t *testing.T) {
	b.Reset()
	h.SetLevel(DEBUG)
	Debug("%d, %s", 1, "OK")
	h.SetLevel(INFO)
	Debug("%d, %s", 1, "OK")
	if b.Len() != 34 {
		t.Fail()
	}
}

func TestSetLevelString(t *testing.T) {
	b.Reset()
	h.SetLevel(DEBUG)
	Debug("%d, %s", 1, "OK")
	h.SetLevelString("info")
	Debug("%d, %s", 1, "OK")
	if b.Len() != 34 {
		t.Fail()
	}
}

func TestSetLevelRange(t *testing.T) {
	b.Reset()
	h.SetLevel(DEBUG)
	Debug("%d, %s", 1, "OK")
	Info("%d, %s", 1, "OK")
	Error("%d, %s", 1, "OK")
	h.SetLevelRange(INFO, WARNING)
	Debug("%d, %s", 1, "OK")
	Info("%d, %s", 1, "OK")
	Error("%d, %s", 1, "OK")
	if b.Len() != 34*4 {
		t.Fail()
	}
	h.lRange = nil
}

func TestSetLevelRangeString(t *testing.T) {
	b.Reset()
	h.SetLevel(DEBUG)
	Debug("%d, %s", 1, "OK")
	Info("%d, %s", 1, "OK")
	Error("%d, %s", 1, "OK")
	h.SetLevelRangeString("INFO", "WARNING")
	Debug("%d, %s", 1, "OK")
	Info("%d, %s", 1, "OK")
	Error("%d, %s", 1, "OK")
	if b.Len() != 34*4 {
		t.Fail()
	}
	h.lRange = nil
}

func TestSetTimeLayout(t *testing.T) {
	b.Reset()
	h.SetLevel(DEBUG)
	h.SetTimeLayout("2006/01/02-15:04:05")
	Error("%d, %s", 1, "OK")
	if b.Len() != 34 {
		t.Fail()
	}
	h.SetTimeLayout(DefaultTimeLayout)
}

func TestSetFilter(t *testing.T) {
	b.Reset()
	h.SetLevel(DEBUG)
	h.SetFilter(func(rd *Record) bool {
		return strings.Contains(rd.Message, "OK")
	})
	Error("%d, %s", 1, "OK")
	if b.Len() != 0 {
		t.Fail()
	}
	h.SetFilter(nil)
}

func TestLoggerHandlerName(t *testing.T) {
	b.Reset()
	DefaultLogger.Name = "DefaultLogger"
	h.SetFormat(func(name, timeString string, rd *Record) string {
		return "[" + timeString + "] " + rd.LoggerName + "." + name + " " + rd.Level.String() + " " + rd.Message + "\n"
	})
	Error("%d, %s", 1, "OK")
	if b.Len() != 49 {
		t.Fail()
	}
	h.SetFormat(DefaultFormat)
}

func BenchmarkEmit(bench *testing.B) {
	h.SetLevel(DEBUG)
	for i := 0; i < bench.N; i++ {
		b.Reset()
		rd := &Record{
			Time:       time.Now(),
			Level:      INFO,
			Message:    fmt.Sprintf("%s, %s", "Hello", "world!"),
			LoggerName: "R",
		}
		h.Emit("SS", rd)
	}
}
