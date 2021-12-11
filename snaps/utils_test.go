package snaps

import (
	"strings"
	"sync"
	"testing"
)

type textScenario struct {
	description string
	expected    string
	formatFunc  func(string) string
}

func TestUtils(t *testing.T) {
	for _, scenario := range []textScenario{
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

			if txt := s.formatFunc("hello world"); txt != s.expected {
				t.Errorf("expected %s - received %s", s.expected, txt)
			}
		})
	}

	t.Run("should create a string from multiple params", func(t *testing.T) {
		expected := "test\nint(5)\nmap[string]string{\"test\":\"test\"}\n"

		if res := takeSnapshot(&[]interface{}{"test", 5, map[string]string{"test": "test"}}); res != expected {
			t.Errorf("not expected snapshot %s", res)
		}
	})

	t.Run("getTestID should increment id on each call [concurrent safe]", func(t *testing.T) {
		wg := sync.WaitGroup{}

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				getTestID("test", "/file")
				wg.Done()
			}()
		}

		wg.Wait()

		testID := getTestID("test", "/file")
		testV2ID := getTestID("test-v2", "/file")

		if testID != "[test - 6]" {
			t.Errorf("wrong test id %s - expected: [test - 6]", testID)
		}

		if testV2ID != "[test-v2 - 1]" {
			t.Errorf("wrong test id %s - expected: [test-v2 - 1]", testV2ID)
		}
	})

	t.Run("dynamicTestIDRegexp", func(t *testing.T) {
		t.Run("should escape regular expressions from testID", func(t *testing.T) {
			escapedTestIDRegexp := dynamicTestIDRegexp(`^\s+$-test`)
			expectedRegexp := `(?:\^\\s\+\$-test[\s\S])(.*[\s\S]*?)(?:---)`

			if expectedRegexp != escapedTestIDRegexp.String() {
				t.Errorf("wrong regex %s - expected %s", escapedTestIDRegexp.String(), expectedRegexp)
			}
		})
	})

	t.Run("should return correct baseCaller", func(t *testing.T) {
		var file string
		expected := "/snaps/utils_test.go"
		func() {
			file, _ = baseCaller()
		}()

		if !strings.Contains(file, expected) {
			t.Errorf("wrong baseCaller %s - expected %s", file, expected)
		}
	})
}
