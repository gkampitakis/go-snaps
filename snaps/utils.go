package snaps

import (
	"bufio"
	"errors"
	"io"
	"math"
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
	updateVAR       = os.Getenv("UPDATE_SNAPS")
	shouldClean     = updateVAR == "true" || updateVAR == "clean"
	defaultConfig   = Config{
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
		Skip(...any)
		Skipf(string, ...any)
		SkipNow()
		Name() string
		Error(...any)
		Log(...any)
		Cleanup(func())
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
func baseCaller(skip int) (string, int) {
	var (
		pc             uintptr
		file, prevFile string
		line, prevLine int
		ok             bool
	)

	for i := skip + 1; ; i++ {
		prevLine = line
		prevFile = file
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

// snapshotScanner returns a new *bufio.Scanner with a `MaxScanTokenSize == math.MaxInt` to read from r.
func snapshotScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, math.MaxInt)
	return s
}

// shouldUpdate determines whether snapshots should be updated
func shouldUpdate(u *bool) bool {
	if isCI {
		return false
	}

	if u != nil {
		return *u
	}

	return updateVAR == "true"
}
