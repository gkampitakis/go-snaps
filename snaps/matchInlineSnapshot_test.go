package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestMatchInlineSnapshot(t *testing.T) {
	t.Run("should error in case of different input from inline snapshot", func(t *testing.T) {
		mockT := test.NewMockTestingT(t)

		mockT.MockError = func(a ...any) {
			test.Equal(
				t,
				"\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 1\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ Received + 0\x1b[0m\n\n\x1b[48;5;225m\x1b"+
					"[38;5;52m- \x1b[0m\x1b[48;5;127m\x1b[38;5;255mdifferent \x1b[0m\x1b[48;5;225m\x1b[38;5;52mvalue\x1b[0m\n\x1b[48;5;159m\x1b"+
					"[38;5;22m+ \x1b[0m\x1b[48;5;159m\x1b[38;5;22mvalue\x1b[0m\n\n",
				a[0].(string),
			)
		}

		MatchInlineSnapshot(mockT, "value", Inline("different value"))

		test.Equal(t, 1, testEvents.items[erred])
	})
}

func TestGetInlineStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "\"hello\"",
		},
		{
			name:     "string with double quotes",
			input:    "quotes: \"test\"",
			expected: "`quotes: \"test\"`",
		},
		{
			name:     "string with single quotes",
			input:    "quotes: 'test'",
			expected: `"quotes: 'test'"`,
		},
		{
			name:     "string with backticks",
			input:    "backtick: `test`",
			expected: `"backtick: ` + "`test`" + `"`,
		},
		{
			name:     "multiline string",
			input:    "line1\nline2",
			expected: "`line1\nline2`",
		},
		{
			name:     "string with special characters",
			input:    "special chars: \t\n\r",
			expected: "`special chars: \t\n\r`",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "\"\"",
		},
		{
			name:     "string with both backticks and newlines",
			input:    "code: `test`\nline2",
			expected: "`code: `test`\nline2`",
		},
		{
			name:     "string with both backticks and double quotes",
			input:    "mixed: `backtick` and \"quote\"",
			expected: "`mixed: `backtick` and \"quote\"`",
		},
		{
			name:     "string with backslashes",
			input:    "path: C:\\Users\\test",
			expected: "`path: C:\\Users\\test`",
		},
		{
			name:     "string with unicode emoji",
			input:    "emoji: 🎉",
			expected: `"emoji: 🎉"`,
		},
		{
			name:     "string with windows line endings",
			input:    "line1\r\nline2",
			expected: "`line1\r\nline2`",
		},
		{
			name:     "string starting with quote",
			input:    "\"starts with quote",
			expected: "`\"starts with quote`",
		},
		{
			name:     "string ending with backtick",
			input:    "ends with backtick`",
			expected: "\"ends with backtick`\"",
		},
		{
			name:     "string with multiple consecutive backticks",
			input:    "code: ``` test ```",
			expected: "\"code: ``` test ```\"",
		},
		{
			name:     "string with only whitespace",
			input:    "   ",
			expected: `"   "`,
		},
		{
			name:     "string with carriage return only",
			input:    "line1\rline2",
			expected: "`line1\rline2`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInlineStringValue(tt.input)
			test.Equal(t, tt.expected, got)
		})
	}
}
