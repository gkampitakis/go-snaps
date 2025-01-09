package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestBaseCallerNested(t *testing.T) {
	file := baseCaller(0)

	test.Contains(t, file, "/snaps/utils_test.go")
}

func testBaseCallerNested(t *testing.T) {
	file := baseCaller(0)

	test.Contains(t, file, "/snaps/utils_test.go")
}

func TestBaseCallerHelper(t *testing.T) {
	t.Helper()
	file := baseCaller(0)

	test.Contains(t, file, "/snaps/utils_test.go")
}

func TestBaseCaller(t *testing.T) {
	t.Run("should return correct baseCaller", func(t *testing.T) {
		var file string

		func() {
			file = baseCaller(1)
		}()

		test.Contains(t, file, "/snaps/utils_test.go")
	})

	t.Run("should return parent function", func(t *testing.T) {
		testBaseCallerNested(t)
	})

	t.Run("should return function's name", func(t *testing.T) {
		TestBaseCallerNested(t)
	})
}

func TestSetCustomCI(t *testing.T) {
	t.Run("should override isCI variable", func(t *testing.T) {
		// Set custom CI to true
		SetCustomCI(true)
		test.True(t, isCI)

		// Set custom CI to false
		SetCustomCI(false)
		test.False(t, isCI)
	})
}
