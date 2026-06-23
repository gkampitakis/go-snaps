package snaps

import (
	"strings"
	"testing"
)

func TestScanLines_CRLF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "LF only",
			input:    "line1\nline2\nline3\n",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "CRLF lines",
			input:    "line1\r\nline2\r\nline3\r\n",
			expected: []string{"line1\r", "line2\r", "line3\r"},
		},
		{
			name:     "Mixed endings",
			input:    "line1\r\nline2\nline3\r\nline4\n",
			expected: []string{"line1\r", "line2", "line3\r", "line4"},
		},
		{
			name:     "No final newline",
			input:    "line1\r\nline2",
			expected: []string{"line1\r", "line2"},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := snapshotScanner(strings.NewReader(tt.input))
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				t.Fatalf("unexpected scan error: %v", err)
			}
			if len(lines) != len(tt.expected) {
				t.Fatalf("expected %d lines, got %d: %v", len(tt.expected), len(lines), lines)
			}
			for i := range lines {
				if lines[i] != tt.expected[i] {
					t.Errorf("line %d: expected %q, got %q", i, tt.expected[i], lines[i])
				}
			}
		})
	}
}
