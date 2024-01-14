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
	MockSkip    func(...any)
	MockSkipf   func(string, ...any)
	MockSkipNow func()
	MockError   func(...any)
	MockLog     func(...any)
	MockCleanup func(func())
}

func (m MockTestingT) Error(args ...any) {
	m.MockError(args...)
}

func (m MockTestingT) Helper() {
	m.MockHelper()
}

func (m MockTestingT) Skip(args ...any) {
	m.MockSkip(args...)
}

func (m MockTestingT) Skipf(format string, args ...any) {
	m.MockSkipf(format, args...)
}

func (m MockTestingT) SkipNow() {
	m.MockSkipNow()
}

func (m MockTestingT) Name() string {
	return m.MockName()
}

func (m MockTestingT) Log(args ...any) {
	m.MockLog(args...)
}

func (m MockTestingT) Cleanup(f func()) {
	m.MockCleanup(f)
}

// Equal asserts expected and received have deep equality
func Equal[A any](t *testing.T, expected, received A) {
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

func CreateTempFile(t *testing.T, data string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "mock.file")

	_ = os.WriteFile(path, []byte(data), os.ModePerm)

	return path
}

// GetFileContent returns the contents of a file
//
// it errors if file doesn't exist
func GetFileContent(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(name)
	if err != nil {
		t.Error(err)
	}

	return string(content)
}

func True(t *testing.T, val bool) {
	t.Helper()

	if !val {
		t.Error("expected true but got false")
	}
}

func False(t *testing.T, val bool) {
	t.Helper()

	if val {
		t.Error("expected false but got true")
	}
}

func Nil(t *testing.T, val any) {
	t.Helper()
	v := reflect.ValueOf(val)

	if val != nil && !v.IsNil() {
		t.Errorf("expected nil but got %v", val)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error but got %s", err)
	}
}
