package snaps

import (
	"sync"
	"testing"
)

func TestUtils(t *testing.T) {
	for _, scenario := range []struct {
		description string
		expected    string
		formatFunc  func(string) string
	}{
		{
			description: "should wrap text with red background",
			expected:    "\u001b[41m\u001b[37;1mhello world\u001b[0m",
			formatFunc:  redBG,
		},
		{
			description: "should wrap text with green background",
			expected:    "\u001b[42m\u001b[37;1mhello world\u001b[0m",
			formatFunc:  greenBG,
		},
		{
			description: "should wrap text with dim text color",
			expected:    "\u001b[2mhello world\u001b[0m",
			formatFunc:  dimText,
		},
		{
			description: "should wrap text with green text color",
			expected:    "\u001b[32;1mhello world\u001b[0m",
			formatFunc:  greenText,
		},
		{
			description: "should wrap text with red text color",
			expected:    "\u001b[31;1mhello world\u001b[0m",
			formatFunc:  redText,
		},
		{
			description: "should wrap text with yellow text color",
			expected:    "\u001b[33;1mhello world\u001b[0m",
			formatFunc:  yellowText,
		},
	} {
		s := scenario

		t.Run(s.description, func(t *testing.T) {
			t.Parallel()

			Equal(t, s.expected, s.formatFunc("hello world"))
		})
	}

	t.Run("should create a string from multiple params", func(t *testing.T) {
		expected := "test\nint(5)\nmap[string]string{\"test\":\"test\"}\n"
		received := takeSnapshot([]interface{}{"test", 5, map[string]string{"test": "test"}})

		Equal(t, expected, received)
	})

	t.Run("getTestID should increment id on each call [concurrent safe]", func(t *testing.T) {
		wg := sync.WaitGroup{}
		registry := newRegistry()

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				registry.getTestID("test", "/file")
				wg.Done()
			}()
		}

		wg.Wait()

		Equal(t, "[test - 6]", registry.getTestID("test", "/file"))
		Equal(t, "[test-v2 - 1]", registry.getTestID("test-v2", "/file"))
	})

	t.Run("dynamicTestIDRegexp", func(t *testing.T) {
		t.Run("should escape regular expressions from testID", func(t *testing.T) {
			escapedTestIDRegexp := dynamicTestIDRegexp(`^\s+$-test`)
			expectedRegexp := `(?:\^\\s\+\$-test[\s\S])(.*[\s\S]*?)(?:---)`

			Equal(t, expectedRegexp, escapedTestIDRegexp.String())
		})
	})

	t.Run("should return correct baseCaller", func(t *testing.T) {
		var (
			file  string
			fName string
		)

		func() {
			file, fName = baseCaller()
		}()

		Contains(t, file, "/snaps/utils_test.go")
		Contains(t, fName, "testing.tRunner")
	})
}
