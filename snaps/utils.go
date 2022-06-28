package snaps

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sync"

	"github.com/gkampitakis/ciinfo"
	"github.com/kr/pretty"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	testsRegistry   = newRegistry()
	snapshotMap     = newSnapshotMap()
	errSnapNotFound = errors.New("snapshot not found")
	_m              = sync.Mutex{}
	isCI            = ciinfo.IsCI
	// Matches [ Test... - number ] testIDs
	testIDRegexp = regexp.MustCompile(`(?m)^\[(Test.* - \d)\]$`)
	spacesRegexp = regexp.MustCompile(`^\s+$`)
	dmp          = diffmatchpatch.New()
	shouldUpdate = getEnvBool("UPDATE_SNAPS", false) && !isCI
)

const (
	resetCode   = "\u001b[0m"
	redBGCode   = "\u001b[41m\u001b[37;1m"
	greenBGCode = "\u001b[42m\u001b[37;1m"
	dimCode     = "\u001b[2m"
	greenCode   = "\u001b[32;1m"
	redCode     = "\u001b[31;1m"
	yellowCode  = "\u001b[33;1m"
	arrowPoint  = "› "
	bulletPoint = "• "
	newLine     = "\n"
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

type printerF func(format string, a ...interface{}) (int, error)

/*
	We track occurrence as in the same test we can run multiple snapshots
	This also helps with keeping track with obsolete snaps
	map[snap path]: map[testname]: <number of snapshots>
*/
type syncRegistry struct {
	values map[string]map[string]int
	sync.Mutex
}

type syncSnapshotMap struct {
	snapshots map[string]map[string]string
	sync.RWMutex
}

// TODO: better naming
func newSnapshotMap() *syncSnapshotMap {
	return &syncSnapshotMap{
		snapshots: map[string]map[string]string{},
	}
}

func (m *syncSnapshotMap) get(snapPath, testID string) (string, bool) {
	m.RLock()
	defer m.RUnlock()

	v, ok := m.snapshots[snapPath][testID]
	return v, ok
}

func (m *syncSnapshotMap) fileLoaded(snapPath string) bool {
	m.RLock()
	defer m.RUnlock()
	_, loaded := m.snapshots[snapPath]

	return loaded
}

func (m *syncSnapshotMap) setFile(snapPath string, value map[string]string) {
	m.Lock()
	defer m.Unlock()

	m.snapshots[snapPath] = value
}

func (m *syncSnapshotMap) setSnapshot(snapPath, testID, value string) {
	m.Lock()
	defer m.Unlock()

	m.snapshots[snapPath][testID] = value
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

	return fmt.Sprintf("%s - %d", tName, occurrence)
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
	}
}

func newRegistry() *syncRegistry {
	return &syncRegistry{
		values: make(map[string]map[string]int),
	}
}

func redBG(txt string) string {
	return fmt.Sprintf("%s%s%s", redBGCode, txt, resetCode)
}

func greenBG(txt string) string {
	return fmt.Sprintf("%s%s%s", greenBGCode, txt, resetCode)
}

func dimText(txt string) string {
	return fmt.Sprintf("%s%s%s", dimCode, txt, resetCode)
}

func greenText(txt string) string {
	return fmt.Sprintf("%s%s%s", greenCode, txt, resetCode)
}

func redText(txt string) string {
	return fmt.Sprintf("%s%s%s", redCode, txt, resetCode)
}

func yellowText(txt string) string {
	return fmt.Sprintf("%s%s%s", yellowCode, txt, resetCode)
}

func takeSnapshot(value interface{}) string {
	return pretty.Sprint(value) + newLine
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
