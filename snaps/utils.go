package snaps

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

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

/*
We track occurrence as in the same test we can run multiple snapshots
This also helps with keeping track with obsolete snaps
map[snap path]: map[testname]: <number of snapshots>
*/
type syncRegistry struct {
	values map[string]map[string]int
	sync.Mutex
}

// Returns the id of the test in the snapshot
// Form [<test-name> - <occurrence>]
func (s *syncRegistry) getTestID(tName, snapPath string) string {
	occurrence := 1
	s.Lock()

	if _, exists := s.values[snapPath]; !exists {
		s.values[snapPath] = make(map[string]int)
	}

	if c, exists := s.values[snapPath][tName]; exists {
		occurrence = c + 1
	}

	s.values[snapPath][tName] = occurrence
	s.Unlock()

	return fmt.Sprintf("[%s - %d]", tName, occurrence)
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

func newRegistry() *syncRegistry {
	return &syncRegistry{
		values: make(map[string]map[string]int),
		Mutex:  sync.Mutex{},
	}
}

func takeSnapshot(snap interface{}) string {
	return escapeEndChars(pretty.Sprint(snap) + "\n")
}

// Matches a specific testID
func dynamicTestIDRegexp(testID string) *regexp.Regexp {
	// e.g (?m)(?:\[TestAdd\/Hello_World\/my-test - 1\][\s\S])(.*[\s\S]*?)(?:^---$)
	return regexp.MustCompile(`(?m)(?:` + regexp.QuoteMeta(testID) + `[\s\S])(.*[\s\S]*?)(?:^---$)`)
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

func unescapeEndChars(input string) string {
	return endCharEscapedRegexp.ReplaceAllLiteralString(input, "---")
}

func escapeEndChars(input string) string {
	// This is for making sure a snapshot doesn't contain an ending char
	return endCharRegexp.ReplaceAllLiteralString(input, "/-/-/-/")
}
