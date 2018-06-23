package logging

import (
	"testing"
)

func TestStdoutColorful(t *testing.T) {
	EnableStdout()
	EnableColorful()
	Debug("%d, %s", 1, "OK")
	Info("%d, %s", 1, "OK")
	Warning("%d, %s", 1, "OK")
	Error("%d, %s", 1, "OK")
	DisableColorful()
	DisableStdout()
}
