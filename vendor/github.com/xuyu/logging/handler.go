package logging

import (
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

const (
	DefaultTimeLayout = "2006-01-02 15:04:05"
)

const (
	y1  = `0123456789`
	y2  = `0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789`
	y3  = `0000000000111111111122222222223333333333444444444455555555556666666666777777777788888888889999999999`
	y4  = `0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789`
	mo1 = `000000000111`
	mo2 = `123456789012`
	d1  = `0000000001111111111222222222233`
	d2  = `1234567890123456789012345678901`
	h1  = `000000000011111111112222`
	h2  = `012345678901234567890123`
	mi1 = `000000000011111111112222222222333333333344444444445555555555`
	mi2 = `012345678901234567890123456789012345678901234567890123456789`
	s1  = `000000000011111111112222222222333333333344444444445555555555`
	s2  = `012345678901234567890123456789012345678901234567890123456789`
	ns1 = `0123456789`
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

	ts,_ := formatTimeHeader(rd.Time)
	s := h.format(name, string(ts), rd)
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

func formatTimeHeader(when time.Time) ([]byte, int) {
	y, mo, d := when.Date()
	h, mi, s := when.Clock()
	ns := when.Nanosecond()/1000000
	//len("2006/01/02 15:04:05.123 ")==24
	var buf [24]byte

	buf[0] = y1[y/1000%10]
	buf[1] = y2[y/100]
	buf[2] = y3[y-y/100*100]
	buf[3] = y4[y-y/100*100]
	buf[4] = '/'
	buf[5] = mo1[mo-1]
	buf[6] = mo2[mo-1]
	buf[7] = '/'
	buf[8] = d1[d-1]
	buf[9] = d2[d-1]
	buf[10] = ' '
	buf[11] = h1[h]
	buf[12] = h2[h]
	buf[13] = ':'
	buf[14] = mi1[mi]
	buf[15] = mi2[mi]
	buf[16] = ':'
	buf[17] = s1[s]
	buf[18] = s2[s]
	buf[19] = '.'
	buf[20] = ns1[ns/100]
	buf[21] = ns1[ns%100/10]
	buf[22] = ns1[ns%10]

	buf[23] = ' '

	return buf[0:], d
}
