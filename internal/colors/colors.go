package colors

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	reset = "\x1b[0m"

	RedBg       = "\x1b[48;5;225m"
	GreenBG     = "\x1b[48;5;159m"
	BoldGreenBG = "\x1b[48;5;23m"
	BoldRedBg   = "\x1b[48;5;127m"

	BoldWhite = "\x1b[1;38;5;255m"

	Dim       = "\x1b[2m"
	Greendiff = "\x1b[38;5;22m"
	Reddiff   = "\x1b[38;5;52m"
	Yellow    = "\x1b[33;1m"
	White     = "\x1b[38;5;255m"
	Green     = "\x1b[32;1m"
	Red       = "\x1b[31;1m"
)

var NOCOLOR = isNoColor()

func isNoColor() bool {
	// https://no-color.org (with any value)
	_, noColor := os.LookupEnv("NO_COLOR")
	// hacky way but should be good enough to support diff on vscode output panel
	term := strings.ToLower(os.Getenv("_"))
	return noColor || strings.Contains(term, "visual") || strings.Contains(term, "code")
}

func Sprint(color, s string) string {
	if NOCOLOR {
		return s
	}

	return fmt.Sprintf("%s%s%s", color, s, reset)
}

func Fprint(w io.Writer, color, s string) {
	if NOCOLOR {
		io.WriteString(w, s)
		return
	}

	fmt.Fprintf(w, "%s%s%s", color, s, reset)
}

/** Only used for the pretty_diff */
func FprintEqual(w io.Writer, s string) {
	if NOCOLOR {
		fmt.Fprintf(w, "  %s", s)
		return
	}

	// we use the space here for aligning with insert and delete sign
	fmt.Fprintf(w, "  %s%s%s", Dim, s, reset)
}

func FprintDelete(w io.Writer, s string) {
	if NOCOLOR {
		fmt.Fprintf(w, "- %s", s)
		return
	}

	// this is for mitigating https://unix.stackexchange.com/q/212933
	// couldn't find a better way.
	if hasNewlineSuffix(s) {
		fmt.Fprintf(w, "%s%s- %s%s\n", Reddiff, RedBg, trimSuffix(s), reset)
	} else {
		fmt.Fprintf(w, "%s%s- %s%s", Reddiff, RedBg, s, reset)
	}
}

func FprintInsert(w io.Writer, s string) {
	if NOCOLOR {
		fmt.Fprintf(w, "+ %s", s)
		return
	}

	if hasNewlineSuffix(s) {
		fmt.Fprintf(w, "%s%s+ %s%s\n", Greendiff, GreenBG, trimSuffix(s), reset)
	} else {
		fmt.Fprintf(w, "%s%s+ %s%s", Greendiff, GreenBG, s, reset)
	}
}

func FprintDeleteBold(w io.Writer, s string) {
	if NOCOLOR {
		io.WriteString(w, s)
		return
	}

	fmt.Fprintf(w, "%s%s%s%s", BoldRedBg, White, s, reset)
}

func FprintInsertBold(w io.Writer, s string) {
	if NOCOLOR {
		io.WriteString(w, s)
		return
	}

	fmt.Fprintf(w, "%s%s%s%s", BoldGreenBG, White, s, reset)
}

func FprintRange(w io.Writer, r1, r2 string) {
	if NOCOLOR {
		fmt.Fprintf(w, "@@ -%s +%s @@\n\n", r1, r2)
		return
	}

	fmt.Fprintf(w, "%s@@ -%s +%s @@%s\n\n", Yellow, r1, r2, reset)
}

func FprintBg(w io.Writer, bgColor, color, s string) {
	if NOCOLOR {
		io.WriteString(w, s)
		return
	}

	if hasNewlineSuffix(s) {
		fmt.Fprintf(w, "%s%s%s%s\n", bgColor, color, trimSuffix(s), reset)
	} else {
		fmt.Fprintf(w, "%s%s%s%s", bgColor, color, s, reset)
	}
}

// hasNewlineSuffix checks if the string contains a "\n" at the end.
func hasNewlineSuffix(s string) bool {
	return strings.HasSuffix(s, "\n")
}

// trimSuffix is like strings.TrimSuffix but specific for "\n" char and without checking
// it exists as we already have earlier.
func trimSuffix(s string) string {
	return s[:len(s)-1]
}
