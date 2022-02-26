package snaps

import (
	"os"
	"testing"
)

// NOTE: this was added at 1.17
func setEnv(t *testing.T, key, value string) {
	t.Helper()

	prevVal, exists := os.LookupEnv(key)
	os.Setenv(key, value)

	if exists {
		t.Cleanup(func() {
			os.Setenv(key, prevVal)
		})
	} else {
		t.Cleanup(func() {
			os.Unsetenv(key)
		})
	}
}

func TestEnv(t *testing.T) {
	t.Run("should return true if env var is 'true'", func(t *testing.T) {
		setEnv(t, "MOCK_ENV", "true")

		res := getEnvBool("MOCK_ENV", false)

		if !res {
			t.Error("getEnvBool should return true")
		}
	})

	t.Run("should return false", func(t *testing.T) {
		setEnv(t, "MOCK_ENV", "")

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
