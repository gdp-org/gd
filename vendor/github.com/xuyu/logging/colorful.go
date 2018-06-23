package logging

type color uint16

const (
	green  = color(0x0002)
	red    = color(0x0004)
	yellow = color(0x000E)
)

func colorful(level logLevel) {
	switch level {
	case ERROR:
		changeColor(red)
	case WARNING:
		changeColor(yellow)
	case INFO:
		changeColor(green)
	}
}
