package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

// This file is "sensitive" to line changes.

func TestBaseCallerNested(t *testing.T) {
	file, line := baseCaller(0)

	test.Contains(t, file, "snaps/basecaller_test.go")
	test.Equal(t, 12, line)
}

func testBaseCallerNested(t *testing.T) {
	file, line := baseCaller(0)

	test.Contains(t, file, "snaps/basecaller_test.go")
	test.Equal(t, 19, line)
}

func TestBaseCallerHelper(t *testing.T) {
	t.Helper()
	file, line := baseCaller(0)

	test.Contains(t, file, "snaps/basecaller_test.go")
	test.Equal(t, 27, line)
}

func TestBaseCaller(t *testing.T) {
	t.Run("should return correct baseCaller", func(t *testing.T) {
		var file string
		var line int

		func() {
			file, line = baseCaller(1)
		}()

		test.Contains(t, file, "snaps/basecaller_test.go")
		test.Equal(t, 40, line)
	})

	t.Run("should return parent function", func(t *testing.T) {
		testBaseCallerNested(t)
	})

	t.Run("should return function's name", func(t *testing.T) {
		TestBaseCallerNested(t)
	})
}
