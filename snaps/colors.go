package snaps

import (
	"fmt"
	"io"
	"strings"
)

const (
	reset = "\x1b[0m"

	redBG       = "\x1b[48;5;225m"
	greenBG     = "\x1b[48;5;159m"
	boldGreenBG = "\x1b[48;5;23m"
	boldRedBG   = "\x1b[48;5;127m"

	dim       = "\x1b[2m"
	greendiff = "\x1b[38;5;22m"
	reddiff   = "\x1b[38;5;52m"
	yellow    = "\x1b[33;1m"
	white     = "\x1b[38;5;255m"
	green     = "\x1b[32;1m"
)

func sprintColored(color, s string) string {
	return fmt.Sprintf("%s%s%s", color, s, reset)
}

func fprintColored(w io.Writer, color, s string) {
	fmt.Fprintf(w, "%s%s%s", color, s, reset)
}

/** Only used for the pretty_diff */
func fprintEqual(w io.Writer, s string) {
	fmt.Fprintf(w, "  %s%s%s", dim, s, reset)
}

func fprintDelete(w io.Writer, s string) {
	// this is for mitigating https://unix.stackexchange.com/q/212933
	// couldn't find a better way.
	if hasNewlineSuffix(s) {
		fmt.Fprintf(w, "%s%s- %s%s\n", reddiff, redBG, trimSuffix(s), reset)
	} else {
		fmt.Fprintf(w, "%s%s- %s%s", reddiff, redBG, s, reset)
	}
}

func fprintInsert(w io.Writer, s string) {
	if hasNewlineSuffix(s) {
		fmt.Fprintf(w, "%s%s+ %s%s\n", greendiff, greenBG, trimSuffix(s), reset)
	} else {
		fmt.Fprintf(w, "%s%s+ %s%s", greendiff, greenBG, s, reset)
	}
}

func fprintDeleteBold(w io.Writer, s string) {
	fmt.Fprintf(w, "%s%s%s%s", boldRedBG, white, s, reset)
}

func fprintInsertBold(w io.Writer, s string) {
	fmt.Fprintf(w, "%s%s%s%s", boldGreenBG, white, s, reset)
}

func fprintRange(w io.Writer, r1, r2 string) {
	fmt.Fprintf(w, "%s@@ -%s +%s @@%s\n\n", yellow, r1, r2, reset)
}

func fprintBgColored(w io.Writer, bgColor, color, s string) {
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
