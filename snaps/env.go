package snaps

import (
	"os"
)

func getEnvBool(variable string, fallback bool) bool {
	e := os.Getenv(variable)
	if e == "" {
		return fallback
	}

	return e == "true"
}
