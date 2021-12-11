package snaps

import (
	"testing"
)

var skippedTests = newSyncSlice()

// Wrapper of testing.Skip
func Skip(t *testing.T, args ...interface{}) {
	t.Helper()

	skippedTests.append(t.Name())
	t.Skip(args...)
}

// Wrapper of testing.Skipf
func Skipf(t *testing.T, format string, args ...interface{}) {
	t.Helper()

	skippedTests.append(t.Name())
	t.Skipf(format, args...)
}

// Wrapper of testing.SkipNow
func SkipNow(t *testing.T) {
	t.Helper()

	skippedTests.append(t.Name())
	t.SkipNow()
}
