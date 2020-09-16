/**
 * Copyright 2019 doglog Author. All rights reserved.
 * Author: Chuck1024
 */

package dlog

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var stdout io.Writer = os.Stdout

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	closeOnce   sync.Once
	format      string
	w           chan *LogRecord
	formatCache formatCacheType
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	consoleWriter := &ConsoleLogWriter{
		format: "%L	%D %T	%M	%S",
		w:           make(chan *LogRecord, LogBufferLength),
		formatCache: formatCacheType{},
	}
	go consoleWriter.run(stdout)
	return consoleWriter
}
func (c *ConsoleLogWriter) SetFormat(format string) {
	c.format = format
}
func (c ConsoleLogWriter) run(out io.Writer) {
	for rec := range c.w {
		fmt.Fprint(out, FormatLogRecord(&c.formatCache, c.format, rec))
	}
}

// This is the ConsoleLogWriter's output method.  This will block if the output
// buffer is full.
func (c ConsoleLogWriter) LogWrite(rec *LogRecord) {
	select {
	case c.w <- rec:
	default:
		select {
		case c.w <- rec:
		case <-time.After(2 * time.Millisecond):
			//add "fatal " prefix in stderr log to trigger sms alert
			fmt.Fprintf(os.Stderr, "fatal term log channel blocked!%v\n", rec)
		}
	}
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (c ConsoleLogWriter) Close() {
	c.closeOnce.Do(func() {
		close(c.w)
		time.Sleep(50 * time.Millisecond) // Try to give console I/O time to complete
	})
}
