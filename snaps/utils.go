package snaps

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"sync"

	"github.com/gkampitakis/ciinfo"
)

var (
	errSnapNotFound = errors.New("snapshot not found")
	isCI            = ciinfo.IsCI
	updateVAR       = os.Getenv("UPDATE_SNAPS")
	shouldClean     = updateVAR == "always" || (updateVAR == "true" && !isCI) ||
		(updateVAR == "clean" && !isCI)
	defaultConfig = Config{
		snapsDir: "__snapshots__",
	}
	isTrimBathBuild = trimPathBuild()
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
	s.Split(scanLines)
	s.Buffer([]byte{}, math.MaxInt)
	return s
}

// shouldUpdate determines whether snapshots should be updated
func shouldUpdate(u *bool) bool {
	if updateVAR == "always" {
		return true
	}

	if isCI {
		return false
	}

	if u != nil {
		return *u
	}

	return updateVAR == "true"
}

// code taken from bufio/scan.go, modified to not terminal \r from the data.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// shouldCreate determines whether snapshots should be created
func shouldCreate(u *bool) bool {
	if updateVAR == "always" {
		return true
	}

	if isCI {
		return false
	}

	if u != nil {
		return *u
	}

	return true
}

// trimPathBuild checks if the build has trimpath setting true
func trimPathBuild() bool {
	keys := []string{"-trimpath", "--trimpath"}
	goFlags := strings.Split(os.Getenv("GOFLAGS"), " ")

	for _, flag := range goFlags {
		if slices.Contains(keys, flag) {
			return true
		}
	}

	bInfo, ok := debug.ReadBuildInfo()
	if ok && len(bInfo.Settings) > 0 {
		for _, info := range bInfo.Settings {
			if slices.Contains(keys, info.Key) {
				return info.Value == "true"
			}
		}
	}

	return runtime.GOROOT() == ""
}
