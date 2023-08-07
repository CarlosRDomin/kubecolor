package color

import (
	"fmt"
)

type Color int

const escape = "\x1b"

const (
	Black Color = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Green256  = 34
	Yellow256 = 226
	DarkGreen = 28
)

func (c Color) sequence() int {
	return int(c)
}

func Apply(val string, c Color) string {
	return fmt.Sprintf("%s[38;5;%dm%s%s[0m", escape, c.sequence(), val, escape)
}
