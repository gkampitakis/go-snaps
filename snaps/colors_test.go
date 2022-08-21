package snaps

import (
	"io"
	"strings"
	"testing"
)

func TestPrintColors(t *testing.T) {
	t.Run("string utils", func(t *testing.T) {
		t.Run("hasNewlineSuffix", func(t *testing.T) {
			Equal(t, false, hasNewlineSuffix("hello \n world"))
			Equal(t, true, hasNewlineSuffix("hello \n world\n"))
		})

		t.Run("trimSuffix", func(t *testing.T) {
			Equal(t, "hello \n world\n", trimSuffix("hello \n world\n\n"))
		})
	})

	t.Run("[sprintColored] should return color wrapped text", func(t *testing.T) {
		Equal(t, "\x1b[2mhello world\x1b[0m", sprintColored(dim, "hello world"))
	})

	for _, v := range []struct {
		name      string
		expected  string
		formatter func(w io.Writer)
	}{
		{
			name:     "[fprintColored] should print string as is",
			expected: "\x1b[2mhello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintColored(w, dim, "hello \n world")
			},
		},
		{
			name:     "[fprintEqual] should print string with dim and space",
			expected: "  \x1b[2mhello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintEqual(w, "hello \n world")
			},
		},
		{
			name:     "[fprintDelete] should return red colored text and escape suffix newline",
			expected: "\x1b[38;5;52m\x1b[48;5;225m- hello \n world\x1b[0m\n",
			formatter: func(w io.Writer) {
				fprintDelete(w, "hello \n world\n")
			},
		},
		{
			name:     "[fprintDelete] should print colored string as is",
			expected: "\x1b[38;5;52m\x1b[48;5;225m- hello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintDelete(w, "hello \n world")
			},
		},
		{
			name:     "[fprintInsert] should return green colored text and escape suffix newline",
			expected: "\x1b[38;5;22m\x1b[48;5;159m+ hello \n world\x1b[0m\n",
			formatter: func(w io.Writer) {
				fprintInsert(w, "hello \n world\n")
			},
		},
		{
			name:     "[fprintInsert] should print colored string as is",
			expected: "\x1b[38;5;22m\x1b[48;5;159m+ hello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintInsert(w, "hello \n world")
			},
		},
		{
			name:     "[fprintBgColored] should return green colored text and escape suffix newline",
			expected: "bg_colorcolorhello \n world\x1b[0m\n",
			formatter: func(w io.Writer) {
				fprintBgColored(w, "bg_color", "color", "hello \n world\n")
			},
		},
		{
			name:     "[fprintBgColored] should print colored string as is",
			expected: "bg_colorcolorhello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintBgColored(w, "bg_color", "color", "hello \n world")
			},
		},
		{
			name:     "[fprintDeleteBold] should apply red bold coloring",
			expected: "\x1b[48;5;127m\x1b[38;5;255mhello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintDeleteBold(w, "hello \n world")
			},
		},
		{
			name:     "[fprintInsertBold] should apply green bold coloring",
			expected: "\x1b[48;5;23m\x1b[38;5;255mhello \n world\x1b[0m",
			formatter: func(w io.Writer) {
				fprintInsertBold(w, "hello \n world")
			},
		},
		{
			name:     "[fprintRange] should apply yellow color and format",
			expected: "\x1b[33;1m@@ -1 +2 @@\x1b[0m\n\n",
			formatter: func(w io.Writer) {
				fprintRange(w, "1", "2")
			},
		},
	} {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			var s strings.Builder
			v.formatter(&s)

			Equal(t, v.expected, s.String())
		})
	}
}
