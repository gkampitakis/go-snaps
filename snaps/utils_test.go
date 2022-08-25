package snaps

import (
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps/internal/test"
)

func TestUtils(t *testing.T) {
	t.Run("should create a string from multiple params", func(t *testing.T) {
		expected := "test\nint(5)\nmap[string]string{\"test\":\"test\"}\n"
		received := takeSnapshot([]interface{}{"test", 5, map[string]string{"test": "test"}})

		test.Equal(t, expected, received)
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

		test.Equal(t, "[test - 6]", registry.getTestID("test", "/file"))
		test.Equal(t, "[test-v2 - 1]", registry.getTestID("test-v2", "/file"))
	})

	t.Run("dynamicTestIDRegexp", func(t *testing.T) {
		t.Run("should escape regular expressions from testID", func(t *testing.T) {
			escapedTestIDRegexp := dynamicTestIDRegexp(`^\s+$-test`)
			expectedRegexp := `(?m)(?:\^\\s\+\$-test[\s\S])(.*[\s\S]*?)(?:^---$)`

			test.Equal(t, expectedRegexp, escapedTestIDRegexp.String())
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

		test.Contains(t, file, "/snaps/utils_test.go")
		test.Contains(t, fName, "testing.tRunner")
	})
}
