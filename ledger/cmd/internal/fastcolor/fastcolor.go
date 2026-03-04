// Package fastcolor is an extreme subset of the fatih/color package to get
// ANSI colors on standard output.
//
// Modified to output the color string to a StringWriter with fixed-width
// formatting (spaces for padding). Minimal color and attribute support.
package fastcolor

import (
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-isatty"
)

type Color string

const (
	Reset     Color = "0"
	Bold      Color = "1"
	FgBlack   Color = "30"
	FgRed     Color = "31"
	FgGreen   Color = "32"
	FgYellow  Color = "33"
	FgBlue    Color = "34"
	FgMagenta Color = "35"
	FgCyan    Color = "36"
	FgWhite   Color = "37"
)

var spaceStr string = strings.Repeat(" ", 132)

// NoColor defines if the output is colorized or not. It's dynamically set to
// false or true based on the stdout's file descriptor referring to a terminal
// or not. It's also set to true if the NO_COLOR environment variable is
// set (regardless of its value). This is a global option and affects all
// colors.
var NoColor = noColorIsSet() || os.Getenv("TERM") == "dumb" ||
	(!isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()))

// noColorIsSet returns true if the environment variable NO_COLOR is set to a non-empty string.
func noColorIsSet() bool {
	return os.Getenv("NO_COLOR") != ""
}

func (c Color) WriteStringFixed(w io.StringWriter, s string, width int, leftpad bool) {
	if !NoColor {
		w.WriteString("\x1b[")
		w.WriteString(string(c))
		w.WriteString("m")
	}

	l := utf8.RuneCountInString(s)
	spaces := width - l
	if spaces > 0 {
		if leftpad {
			w.WriteString(spaceStr[:spaces])
			w.WriteString(s)
		} else {
			w.WriteString(s)
			w.WriteString(spaceStr[:spaces])
		}
	} else {
		w.WriteString(s[:width])
	}

	if !NoColor {
		w.WriteString("\x1b[0m")
	}
}
