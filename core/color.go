package core

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ANSI color codes
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
)

var colorEnabled = term.IsTerminal(int(os.Stdout.Fd()))

func colorize(color, s string) string {
	if !colorEnabled {
		return s
	}
	return color + s + reset
}

func colorf(color, format string, a ...any) string {
	return colorize(color, fmt.Sprintf(format, a...))
}

// Colorize is the exported version for use outside core package.
func Colorize(color, s string) string { return colorize(color, s) }

// Colorf is the exported version for use outside core package.
func Colorf(color, format string, a ...any) string { return colorf(color, format, a...) }

// Exported color constants for use outside core package.
const (
	Bold    = bold
	Dim     = dim
	Red     = red
	Green   = green
	Yellow  = yellow
	Blue    = blue
	Cyan    = cyan
	White   = white
	Magenta = magenta
)
