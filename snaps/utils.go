package snaps

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sync"

	"github.com/gkampitakis/ciinfo"
	"github.com/kr/pretty"
)

var (
	testsRegistry   = newRegistry()
	errSnapNotFound = errors.New("snapshot not found")
	_m              = sync.Mutex{}
	isCI            = ciinfo.IsCI
	// Matches [ Test... - number ] testIDs
	testIDRegexp         = regexp.MustCompile(`(?m)^\[(Test.* - \d)\]$`)
	endCharRegexp        = regexp.MustCompile(`(?m)(^---$)`)
	endCharEscapedRegexp = regexp.MustCompile(`(?m)(^/-/-/-/$)`)
	shouldUpdate         = getEnvBool("UPDATE_SNAPS", false) && !isCI
)

const (
	arrowPoint  = "› "
	bulletPoint = "• "
	snapsDir    = "__snapshots__"
	snapsExt    = ".snap"
)

type set map[string]struct{}
type testingT interface {
	Helper()
	Skip(args ...interface{})
	Skipf(format string, args ...interface{})
	SkipNow()
	Name() string
	Error(args ...interface{})
	Log(args ...interface{})
}

/*
We track occurrence as in the same test we can run multiple snapshots
This also helps with keeping track with obsolete snaps
map[snap path]: map[testname]: <number of snapshots>
*/
type syncRegistry struct {
	values map[string]map[string]int
	_m     sync.Mutex
}

// Returns the id of the test in the snapshot
// Form [<test-name> - <occurrence>]
func (s *syncRegistry) getTestID(tName, snapPath string) string {
	occurrence := 1
	s._m.Lock()

	if _, exists := s.values[snapPath]; !exists {
		s.values[snapPath] = make(map[string]int)
	}

	if c, exists := s.values[snapPath][tName]; exists {
		occurrence = c + 1
	}

	s.values[snapPath][tName] = occurrence
	s._m.Unlock()

	return fmt.Sprintf("[%s - %d]", tName, occurrence)
}

type syncSlice struct {
	values []string
	_m     sync.Mutex
}

func (s *syncSlice) append(elems ...string) {
	s._m.Lock()
	defer s._m.Unlock()

	s.values = append(s.values, elems...)
}

func newSyncSlice() *syncSlice {
	return &syncSlice{
		values: []string{},
		_m:     sync.Mutex{},
	}
}

func newRegistry() *syncRegistry {
	return &syncRegistry{
		values: make(map[string]map[string]int),
		_m:     sync.Mutex{},
	}
}

func takeSnapshot(objects []interface{}) string {
	var snapshot string

	for i := 0; i < len(objects); i++ {
		snapshot += pretty.Sprint(objects[i]) + "\n"
	}

	return escapeEndChars(snapshot)
}

// Matches a specific testID
func dynamicTestIDRegexp(testID string) *regexp.Regexp {
	// e.g (?m)(?:\[TestAdd\/Hello_World\/my-test - 1\][\s\S])(.*[\s\S]*?)(?:^---$)
	return regexp.MustCompile(`(?m)(?:` + regexp.QuoteMeta(testID) + `[\s\S])(.*[\s\S]*?)(?:^---$)`)
}

// Returns the path where the "user" tests are running and the function name
func baseCaller() (string, string) {
	var (
		ok             bool
		pc             uintptr
		file, prevFile string
		funcName       string
	)

	for i := 0; ; i++ {
		prevFile = file
		pc, file, _, ok = runtime.Caller(i)
		if !ok {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}

		funcName = f.Name()
		if f.Name() == "testing.tRunner" {
			break
		}
	}

	return prevFile, funcName
}

func unescapeEndChars(input string) string {
	return endCharEscapedRegexp.ReplaceAllLiteralString(input, "---")
}

func escapeEndChars(input string) string {
	// This is for making sure a snapshot doesn't contain an ending char
	return endCharRegexp.ReplaceAllLiteralString(input, "/-/-/-/")
}
