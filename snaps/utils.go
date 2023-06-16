package snaps

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/gkampitakis/ciinfo"
)

var (
	errSnapNotFound = errors.New("snapshot not found")
	isCI            = ciinfo.IsCI
	envVar          = os.Getenv("UPDATE_SNAPS")
	shouldUpdate    = envVar == "true" && !isCI
	shouldClean     = shouldUpdate || envVar == "clean" && !isCI
	defaultConfig   = config{
		snapsDir: "__snapshots__",
	}
)

const (
	arrowSymbol   = "› "
	bulletSymbol  = "• "
	errorSymbol   = "✕ "
	successSymbol = "✓ "
	updateSymbol  = "✎ "
	skipSymbol    = "⟳ "
	enterSymbol   = "↳ "
	newLineSymbol = "↵"

	snapsExt    = ".snap"
	endSequence = "---"
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

// Returns the path and the line where the "user" tests are running
func baseCaller(skip int) (string, int) {
	var (
		pc             uintptr
		file, prevFile string
		line, prevLine int
		ok             bool
	)

	for i := skip + 1; ; i++ {
		prevFile = file
		prevLine = line
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			return prevFile, prevLine
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			return prevFile, prevLine
		}

		if f.Name() == "testing.tRunner" {
			return prevFile, prevLine
		}

		if strings.HasSuffix(filepath.Base(file), "_test.go") {
			return file, line
		}
	}
}

// shouldUpdateSingle returns if a single should be updated or not
//
// it depends on the general should update or if the given name on `UPDATE_SNAPS` matches the current test.
func shouldUpdateSingle(tName string) bool {
	return shouldUpdate || (!isCI && envVar != "" && strings.HasPrefix(tName, envVar))
}
