package snaps

import "testing"

func TestEnv(t *testing.T) {
	t.Run("should return true if env var is 'true'", func(t *testing.T) {
		t.Setenv("MOCK_ENV", "true")

		res := getEnvBool("MOCK_ENV", false)

		if !res {
			t.Error("getEnvBool should return true")
		}
	})

	t.Run("should return false", func(t *testing.T) {
		t.Setenv("MOCK_ENV", "")

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
