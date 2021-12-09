package snaps

import (
	"strings"
	"sync"
	"testing"
)

func TestUtils(t *testing.T) {
	t.Run("should wrap text with red background", func(t *testing.T) {
		expectedTxt := "\u001b[41m\u001b[37;1mhello world\u001b[0m"

		if txt := redBG("hello world"); txt != expectedTxt {
			t.Errorf("expected red background %s", txt)
		}
	})

	t.Run("should wrap text with green background", func(t *testing.T) {
		expectedTxt := "\u001b[42m\u001b[37;1mhello world\u001b[0m"

		if txt := greenBG("hello world"); txt != expectedTxt {
			t.Errorf("expected green background %s", txt)
		}
	})

	t.Run("should wrap text with dim text color", func(t *testing.T) {
		expectedTxt := "\u001b[2mhello world\u001b[0m"

		if txt := dimText("hello world"); txt != expectedTxt {
			t.Errorf("expected dim color text %s", txt)
		}
	})

	t.Run("should wrap text with green text color", func(t *testing.T) {
		expectedTxt := "\u001b[32;1mhello world\u001b[0m"

		if txt := greenText("hello world"); txt != expectedTxt {
			t.Errorf("expected green color text %s", txt)
		}
	})

	t.Run("should wrap text with red text color", func(t *testing.T) {
		expectedTxt := "\u001b[31;1mhello world\u001b[0m"

		if txt := redText("hello world"); txt != expectedTxt {
			t.Errorf("expected red color text %s", txt)
		}
	})

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
				getTestID("test")
				wg.Done()
			}()
		}

		wg.Wait()

		testID := getTestID("test")
		test_v2ID := getTestID("test-v2")

		if testID != "[test - 6]" {
			t.Errorf("wrong test id %s - expected: [test - 6]", testID)
		}

		if test_v2ID != "[test-v2 - 1]" {
			t.Errorf("wrong test id %s - expected: [test-v2 - 1]", test_v2ID)
		}
	})

	t.Run("testIDRegex", func(t *testing.T) {
		t.Run("should escape regular expressions from testID", func(t *testing.T) {
			escapedTestIDRegex := testIDRegex(`^\s+$-test`)
			expectedRegex := `(?:\^\\s\+\$-test[\s\S])(.*[\s\S]*?)(?:---)`

			if expectedRegex != escapedTestIDRegex.String() {
				t.Errorf("wrong regex %s - expected %s", escapedTestIDRegex.String(), expectedRegex)
			}
		})
	})

	t.Run("should return correct baseCaller", func(t *testing.T) {
		var file string
		expected := "/snaps/utils_test.go"
		func() {
			file = baseCaller()
		}()

		if !strings.Contains(file, expected) {
			t.Errorf("wrong baseCaller %s - expected %s", file, expected)
		}
	})
}
