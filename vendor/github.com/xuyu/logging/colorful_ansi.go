// +build !windows

package logging

import (
	"os"
)

func resetColorful() {
	_, _ = os.Stdout.WriteString("\x1b[0m")
}

func changeColor(c color) {
	switch c {
	case red:
		_, _ = os.Stdout.WriteString("\x1b[31;1m")
	case yellow:
		_, _ = os.Stdout.WriteString("\x1b[33;1m")
	case green:
		_, _ = os.Stdout.WriteString("\x1b[32;1m")
	}
}
