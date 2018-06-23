package logging

import (
	"bytes"
	"io"
	"os"
	"sync"
)

const (
	DefaultTimeLayout = "2006-01-02 15:04:05"
)

var (
	FileCreateFlag             = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	FileCreatePerm os.FileMode = 0640
)

var DefaultFormat = func(name, timeString string, rd *Record) string {
	return "[" + timeString + "] " + rd.Level.String() + " " + rd.Message + "\n"
}

type Handler struct {
	mutex  sync.Mutex
	buffer *bytes.Buffer
	writer io.Writer
	level  logLevel
	lRange *levelRange
	layout string
	format func(string, string, *Record) string
	filter func(*Record) bool
	before func(*Record, io.ReadWriter)
	after  func(*Record, int64)
}

func NewHandler(out io.Writer) *Handler {
	return &Handler{
		buffer: bytes.NewBuffer(nil),
		writer: out,
		level:  DEBUG,
		layout: DefaultTimeLayout,
		format: DefaultFormat,
	}
}

func (h *Handler) Close() error {
	var err error
	h.mutex.Lock()
	closer, ok := h.writer.(io.Closer)
	if ok {
		err = closer.Close()
	}
	h.writer = nil
	h.mutex.Unlock()
	return err
}

func (h *Handler) SetLevel(level logLevel) {
	h.level = level
}

func (h *Handler) SetLevelString(s string) {
	h.SetLevel(StringToLogLevel(s))
}

func (h *Handler) SetLevelRange(minLevel, maxLevel logLevel) {
	h.lRange = &levelRange{minLevel, maxLevel}
}

func (h *Handler) SetLevelRangeString(smin, smax string) {
	h.SetLevelRange(StringToLogLevel(smin), StringToLogLevel(smax))
}

func (h *Handler) SetTimeLayout(layout string) {
	h.layout = layout
}

func (h *Handler) SetFormat(format func(string, string, *Record) string) {
	h.format = format
}

func (h *Handler) SetFilter(f func(*Record) bool) {
	h.filter = f
}

func (h *Handler) Emit(name string, rd *Record) {
	if h.lRange != nil {
		if !h.lRange.contains(rd.Level) {
			return
		}
	} else if h.level > rd.Level {
		return
	}
	h.handleRecord(name, rd)
}

func (h *Handler) handleRecord(name string, rd *Record) {
	if h.filter != nil && h.filter(rd) {
		return
	}
	s := h.format(name, rd.Time.Format(h.layout), rd)
	h.mutex.Lock()
	if h.writer == nil {
		h.mutex.Unlock()
		return
	}
	if h.before == nil {
		n, err := io.WriteString(h.writer, s)
		if err == nil && h.after != nil {
			h.after(rd, int64(n))
		}
		h.mutex.Unlock()
		return
	}
	h.buffer.Reset()
	_, _ = io.WriteString(h.buffer, s)
	h.before(rd, h.buffer)
	n, err := io.Copy(h.writer, h.buffer)
	if err == nil && h.after != nil {
		h.after(rd, n)
	}
	h.mutex.Unlock()
}
