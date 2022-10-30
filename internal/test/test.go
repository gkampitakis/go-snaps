package test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type MockTestingT struct {
	MockHelper  func()
	MockName    func() string
	MockSkip    func(args ...interface{})
	MockSkipf   func(format string, args ...interface{})
	MockSkipNow func()
	MockError   func(args ...interface{})
	MockLog     func(args ...interface{})
}

func (m MockTestingT) Error(args ...interface{}) {
	m.MockError(args...)
}

func (m MockTestingT) Helper() {
	m.MockHelper()
}

func (m MockTestingT) Skip(args ...interface{}) {
	m.MockSkip(args...)
}

func (m MockTestingT) Skipf(format string, args ...interface{}) {
	m.MockSkipf(format, args...)
}

func (m MockTestingT) SkipNow() {
	m.MockSkipNow()
}

func (m MockTestingT) Name() string {
	return m.MockName()
}

func (m MockTestingT) Log(args ...interface{}) {
	m.MockLog(args...)
}

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

// NOTE: this was added at 1.17
func SetEnv(t *testing.T, key, value string) {
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

func CreateTempFile(t *testing.T, data string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "mock.file")

	_ = os.WriteFile(path, []byte(data), os.ModePerm)

	return path
}
