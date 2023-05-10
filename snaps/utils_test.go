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

func TestShouldUpdateSingle(t *testing.T) {
	setup := func() {
		shouldUpdate = false
		isCI = false
		envVar = ""
	}

	t.Run("should be true if shouldUpdate", func(t *testing.T) {
		setup()

		shouldUpdate = true
		test.True(t, shouldUpdateSingle(""))
	})

	t.Run("should be true if specific test provided from envVar", func(t *testing.T) {
		setup()
		envVar = "mock-test/"
		shouldUpdate = false

		test.True(t, shouldUpdateSingle("mock-test/should_pass - 1"))
		test.False(t, shouldUpdateSingle("mock-test-2/should_pass - 1"))
	})

	t.Run("should be false if running on CI", func(t *testing.T) {
		setup()
		envVar = "mock-test/"
		isCI = true

		test.False(t, shouldUpdateSingle("mock-test/should_pass - 1"))
	})

	t.Run("should be false if not envVar provided", func(t *testing.T) {
		setup()

		test.False(t, shouldUpdateSingle("mock-test"))
	})
}
