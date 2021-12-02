package snaps

import (
	"os"
)

func getEnvBool(variable string, fallback bool) bool {
	e, exists := os.LookupEnv(variable)
	if !exists {
		return fallback
	}

	return e == "true"
}
