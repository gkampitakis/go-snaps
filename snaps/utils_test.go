package snaps

import (
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps/internal/test"
)

func TestUtils(t *testing.T) {
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
}

func TestBaseCallerNested(t *testing.T) {
	file, fName := baseCaller()

	test.Contains(t, file, "/snaps/utils_test.go")
	test.Contains(t, fName, "TestBaseCallerNested")
}

func testBaseCallerNested(t *testing.T) {
	file, fName := baseCaller()

	test.Contains(t, file, "/snaps/utils_test.go")
	test.Contains(t, fName, "TestBaseCaller")
}

func TestBaseCallerHelper(t *testing.T) {
	t.Helper()
	file, fName := baseCaller()

	test.Contains(t, file, "/snaps/utils_test.go")
	test.Contains(t, fName, "TestBaseCaller")
}

func TestBaseCaller(t *testing.T) {
	t.Run("should return correct baseCaller", func(t *testing.T) {
		var (
			file  string
			fName string
		)

		func() {
			file, fName = baseCaller()
		}()

		test.Contains(t, file, "/snaps/utils_test.go")
		test.Contains(t, fName, "TestBaseCaller")
	})

	t.Run("should return parent function", func(t *testing.T) {
		testBaseCallerNested(t)
	})

	t.Run("should return function's name", func(t *testing.T) {
		TestBaseCallerNested(t)
	})
}
