package snaps

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/gkampitakis/ciinfo"
)

var (
	errSnapNotFound = errors.New("snapshot not found")
	isCI            = ciinfo.IsCI
	envVar          = os.Getenv("UPDATE_SNAPS")
	shouldUpdate    = envVar == "true" && !isCI
	shouldClean     = shouldUpdate || envVar == "clean" && !isCI
)

const (
	arrowSymbol   = "› "
	bulletSymbol  = "• "
	errorSymbol   = "✕ "
	successSymbol = "✓ "
	updateSymbol  = "✎ "
	skipSymbol    = "⟳ "
	enterSymbol   = "↳ "

	snapsDir = "__snapshots__"
	snapsExt = ".snap"
)

type (
	set      map[string]struct{}
	testingT interface {
		Helper()
		Skip(args ...interface{})
		Skipf(format string, args ...interface{})
		SkipNow()
		Name() string
		Error(args ...interface{})
		Log(args ...interface{})
	}
)

func (s set) Has(i string) bool {
	_, has := s[i]
	return has
}

type syncSlice struct {
	values []string
	sync.Mutex
}

func (s *syncSlice) append(elems ...string) {
	s.Lock()
	defer s.Unlock()

	s.values = append(s.values, elems...)
}

func newSyncSlice() *syncSlice {
	return &syncSlice{
		values: []string{},
		Mutex:  sync.Mutex{},
	}
}

// Returns the path where the "user" tests are running
func baseCaller(skip int) string {
	var (
		pc             uintptr
		file, prevFile string
	)

	for i := skip + 1; ; i++ {
		prevFile = file
		pc, file, _, _ = runtime.Caller(i)

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}

		funcName := f.Name()
		if funcName == "testing.tRunner" {
			break
		}

		// special case handling test runners
		// tested with testify/suite, packagestest and testcase
		segments := strings.Split(funcName, ".")
		for _, segment := range segments {
			if !isTest(segment, "Test") {
				continue
			}

			// packagestest is same as tRunner where we step one caller further
			// so we need to return the prevFile in testcase and testify/suite we return the current file
			// e.g. funcName golang.org/x/tools/go/packages/packagestest.TestAll.func1
			if strings.Contains(funcName, "packagestest") {
				// return only the Function Name
				// e.g. "go-snaps-testing-suite/src/issues.(*ExampleTestSuite).TestExampleSnapshot"
				// will return TestExampleSnapshot
				return prevFile
			}

			return file
		}
	}

	return prevFile
}

// Stolen from the `go test` tool
//
// isTest tells whether name looks like a test
// It is a Test (say) if there is a character after Test that is not a lower-case letter
func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) { // "Test" is ok
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(r)
}

// shouldUpdateSingle returns if a single should be updated or not
//
// it depends on the general should update or if the given name on `UPDATE_SNAPS` matches the current test.
func shouldUpdateSingle(tName string) bool {
	return shouldUpdate || (!isCI && envVar != "" && strings.HasPrefix(tName, envVar))
}
