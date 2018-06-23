package logging

import (
	"io"
	"os"
)

const (
	stdoutHandlerName = "STDOUT"
)

var (
	StdoutHandler *Handler
)

func init() {
	StdoutHandler = NewHandler(os.Stdout)
	EnableStdout()
	EnableColorful()
}

func DisableStdout() {
	delete(DefaultLogger.Handlers, stdoutHandlerName)
}

func EnableStdout() {
	DefaultLogger.Handlers[stdoutHandlerName] = StdoutHandler
}

func EnableColorful() {
	StdoutHandler.before = func(rd *Record, buf io.ReadWriter) {
		colorful(rd.Level)
	}
	StdoutHandler.after = func(*Record, int64) {
		resetColorful()
	}
}

func DisableColorful() {
	resetColorful()
	StdoutHandler.before = nil
	StdoutHandler.after = nil
}
