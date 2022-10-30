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

func TestEnv(t *testing.T) {
	t.Run("should return true if env var is 'true'", func(t *testing.T) {
		test.SetEnv(t, "MOCK_ENV", "true")

		res := getEnvBool("MOCK_ENV", false)

		if !res {
			t.Error("getEnvBool should return true")
		}
	})

	t.Run("should return false", func(t *testing.T) {
		test.SetEnv(t, "MOCK_ENV", "")

		res := getEnvBool("MOCK_ENV", true)

		if res {
			t.Error("getEnvBool should return false")
		}
	})

	t.Run("should return fallback value for non existing env var", func(t *testing.T) {
		res := getEnvBool("MISSING_ENV", true)

		if !res {
			t.Error("getEnvBool should return false")
		}
	})
}
