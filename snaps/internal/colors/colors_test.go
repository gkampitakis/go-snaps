package colors

import (
	"io"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps/internal/test"
)

func TestPrintColors(t *testing.T) {
	t.Run("string utils", func(t *testing.T) {
		t.Run("hasNewlineSuffix", func(t *testing.T) {
			test.Equal(t, false, hasNewlineSuffix("hello \n world"))
			test.Equal(t, true, hasNewlineSuffix("hello \n world\n"))
		})

		t.Run("trimSuffix", func(t *testing.T) {
			test.Equal(t, "hello \n world\n", trimSuffix("hello \n world\n\n"))
		})
	})

	t.Run("no color", func(t *testing.T) {
		t.Cleanup(func() {
			NOCOLOR = false
		})
		NOCOLOR = true

		t.Run("[Sprint] should return text as is", func(t *testing.T) {
			test.Equal(t, "hello world", Sprint(Dim, "hello world"))
		})

		for _, v := range []struct {
			name      string
			expected  string
			formatter func(w io.Writer)
		}{
			{
				name:     "[Fprint] should print string as is",
				expected: "hello \n world",
				formatter: func(w io.Writer) {
					Fprint(w, Dim, "hello \n world")
				},
			},
			{
				name:     "[FprintEqual] should print string with space",
				expected: "  hello \n world",
				formatter: func(w io.Writer) {
					FprintEqual(w, "hello \n world")
				},
			},
			{
				name:     "[FprintDelete] should print string with - sign",
				expected: "- hello \n world\n",
				formatter: func(w io.Writer) {
					FprintDelete(w, "hello \n world\n")
				},
			},
			{
				name:     "[FprintInsert] should print string with + sign",
				expected: "+ hello \n world\n",
				formatter: func(w io.Writer) {
					FprintInsert(w, "hello \n world\n")
				},
			},
			{
				name:     "[FprintBg] should print string as is",
				expected: "hello \n world\n",
				formatter: func(w io.Writer) {
					FprintBg(w, "bg_color", "color", "hello \n world\n")
				},
			},
			{
				name:     "[FprintDeleteBold] should print string as is",
				expected: "hello \n world",
				formatter: func(w io.Writer) {
					FprintDeleteBold(w, "hello \n world")
				},
			},
			{
				name:     "[FprintInsertBold] should print string as is",
				expected: "hello \n world",
				formatter: func(w io.Writer) {
					FprintInsertBold(w, "hello \n world")
				},
			},
			{
				name:     "[FprintRange] should return range format",
				expected: "@@ -1 +2 @@\n\n",
				formatter: func(w io.Writer) {
					FprintRange(w, "1", "2")
				},
			},
		} {
			v := v
			t.Run(v.name, func(t *testing.T) {
				t.Parallel()

				var s strings.Builder
				v.formatter(&s)

				test.Equal(t, v.expected, s.String())
			})
		}
	})

	t.Run("with color", func(t *testing.T) {
		t.Run("[Sprint] should return color wrapped text", func(t *testing.T) {
			test.Equal(t, "\x1b[2mhello world\x1b[0m", Sprint(Dim, "hello world"))
		})

		for _, v := range []struct {
			name      string
			expected  string
			formatter func(w io.Writer)
		}{
			{
				name:     "[Fprint] should print string as is",
				expected: "\x1b[2mhello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					Fprint(w, Dim, "hello \n world")
				},
			},
			{
				name:     "[FprintEqual] should print string with dim and space",
				expected: "  \x1b[2mhello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					FprintEqual(w, "hello \n world")
				},
			},
			{
				name:     "[FprintDelete] should return red colored text and escape suffix newline",
				expected: "\x1b[38;5;52m\x1b[48;5;225m- hello \n world\x1b[0m\n",
				formatter: func(w io.Writer) {
					FprintDelete(w, "hello \n world\n")
				},
			},
			{
				name:     "[FprintDelete] should print colored string as is",
				expected: "\x1b[38;5;52m\x1b[48;5;225m- hello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					FprintDelete(w, "hello \n world")
				},
			},
			{
				name:     "[FprintInsert] should return green colored text and escape suffix newline",
				expected: "\x1b[38;5;22m\x1b[48;5;159m+ hello \n world\x1b[0m\n",
				formatter: func(w io.Writer) {
					FprintInsert(w, "hello \n world\n")
				},
			},
			{
				name:     "[FprintInsert] should print colored string as is",
				expected: "\x1b[38;5;22m\x1b[48;5;159m+ hello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					FprintInsert(w, "hello \n world")
				},
			},
			{
				name:     "[FprintBg] should return green colored text and escape suffix newline",
				expected: "bg_colorcolorhello \n world\x1b[0m\n",
				formatter: func(w io.Writer) {
					FprintBg(w, "bg_color", "color", "hello \n world\n")
				},
			},
			{
				name:     "[FprintBg] should print colored string as is",
				expected: "bg_colorcolorhello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					FprintBg(w, "bg_color", "color", "hello \n world")
				},
			},
			{
				name:     "[FprintDeleteBold] should apply red bold coloring",
				expected: "\x1b[48;5;127m\x1b[38;5;255mhello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					FprintDeleteBold(w, "hello \n world")
				},
			},
			{
				name:     "[FprintInsertBold] should apply green bold coloring",
				expected: "\x1b[48;5;23m\x1b[38;5;255mhello \n world\x1b[0m",
				formatter: func(w io.Writer) {
					FprintInsertBold(w, "hello \n world")
				},
			},
			{
				name:     "[FprintRange] should apply yellow color and format",
				expected: "\x1b[33;1m@@ -1 +2 @@\x1b[0m\n\n",
				formatter: func(w io.Writer) {
					FprintRange(w, "1", "2")
				},
			},
		} {
			v := v
			t.Run(v.name, func(t *testing.T) {
				t.Parallel()

				var s strings.Builder
				v.formatter(&s)

				test.Equal(t, v.expected, s.String())
			})
		}
	})
}
