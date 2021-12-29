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

type Set map[string]struct{}

type syncMap struct {
	values map[string]map[string]int
	_m     sync.Mutex
}

type syncSlice struct {
	values []string
	_m     sync.Mutex
}

func (s *syncSlice) append(value string) {
	s._m.Lock()
	defer s._m.Unlock()

	s.values = append(s.values, value)
}

func newSyncSlice() *syncSlice {
	return &syncSlice{
		values: []string{},
		_m:     sync.Mutex{},
	}
}

func newSyncMap() *syncMap {
	return &syncMap{
		values: make(map[string]map[string]int),
		_m:     sync.Mutex{},
	}
}

var (
	/*
		We track occurrence as in the same test we can run multiple snapshots
		This also helps with keeping track with obsolete snaps
		map[filepath]: map[testname]: <number of snapshots>
	*/
	testsOccur      = newSyncMap()
	errSnapNotFound = errors.New("snapshot not found")
	_m              = sync.Mutex{}
	shouldUpdate    = getEnvBool("UPDATE_SNAPS", false) && !ciinfo.IsCI
	// Matches [ Test ... ] testIDs
	testIDRegexp = regexp.MustCompile(`^\[([Test].+)]$`)
	spacesRegexp = regexp.MustCompile(`^\s+$`)
	dmp          = diffmatchpatch.New()
)

const (
	resetCode   = "\u001b[0m"
	redBGCode   = "\u001b[41m\u001b[37;1m"
	greenBGCode = "\u001b[42m\u001b[37;1m"
	dimCode     = "\u001b[2m"
	greenCode   = "\u001b[32;1m"
	redCode     = "\u001b[31;1m"
	yellowCode  = "\u001b[33;1m"
)

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

func takeSnapshot(objects *[]interface{}) string {
	var snapshot string

	for i := 0; i < len(*objects); i++ {
		snapshot += pretty.Sprint((*objects)[i]) + "\n"
	}

	return snapshot
}

// Returns the id of the test in the snapshot
// Form [<test-name> - <occurrence>]
func getTestID(tName, fPath string) string {
	occurrence := 1
	testsOccur._m.Lock()

	if _, exists := testsOccur.values[fPath]; !exists {
		testsOccur.values[fPath] = make(map[string]int)
	}

	if c, exists := testsOccur.values[fPath][tName]; exists {
		occurrence = c + 1
	}

	testsOccur.values[fPath][tName] = occurrence
	testsOccur._m.Unlock()

	return fmt.Sprintf("[%s - %d]", tName, occurrence)
}

func dynamicTestIDRegexp(testID string) *regexp.Regexp {
	// e.g (?:\[TestAdd\/Hello_World\/my-test - 1\][\s\S])(.*[\s\S]*?)(?:---)
	return regexp.MustCompile(`(?:` + regexp.QuoteMeta(testID) + `[\s\S])(.*[\s\S]*?)(?:---)`)
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
