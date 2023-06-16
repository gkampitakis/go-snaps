package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestBaseCallerNested(t *testing.T) {
	file, line := baseCaller(0)

	test.Contains(t, file, "/snaps/utils_test.go")
	test.Equal(t, 10, line)
}

func testBaseCallerNested(t *testing.T) {
	file, line := baseCaller(0)

	test.Contains(t, file, "/snaps/utils_test.go")
	test.Equal(t, 17, line)
}

func TestBaseCallerHelper(t *testing.T) {
	t.Helper()
	file, line := baseCaller(0)

	test.Contains(t, file, "/snaps/utils_test.go")
	test.Equal(t, 25, line)
}

func TestBaseCaller(t *testing.T) {
	t.Run("should return correct baseCaller", func(t *testing.T) {
		var file string
		var line int

		func() {
			file, line = baseCaller(1)
		}()

		test.Contains(t, file, "/snaps/utils_test.go")
		test.Equal(t, 38, line)
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
