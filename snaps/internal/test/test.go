package test

import (
	"reflect"
	"strings"
	"testing"
)

// Equal asserts expected and received have deep equality
func Equal(t *testing.T, expected, received interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\n[expected]: %v\n[received]: %v", expected, received)
	}
}

// Contains reports whether a substr is inside s
func Contains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("\n [expected] %s to contain %s", s, substr)
	}
}

// NotCalled is going to mark a test as failed if called
func NotCalled(t *testing.T) {
	t.Helper()
	t.Errorf("function was not expected to be called")
}
