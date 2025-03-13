package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestBaseCallerNested(t *testing.T) {
	file, _ := baseCaller(0)
	t.Error("fix line")

	test.Contains(t, file, "/snaps/utils_test.go")
}

func testBaseCallerNested(t *testing.T) {
	file, _ := baseCaller(0)
	t.Error("fix line")

	test.Contains(t, file, "/snaps/utils_test.go")
}

func TestBaseCallerHelper(t *testing.T) {
	t.Helper()
	file, _ := baseCaller(0)
	t.Error("fix line")

	test.Contains(t, file, "/snaps/utils_test.go")
}

func TestBaseCaller(t *testing.T) {
	t.Run("should return correct baseCaller", func(t *testing.T) {
		var file string

		func() {
			t.Error("fix line")
			file, _ = baseCaller(1)
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
