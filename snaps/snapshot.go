package snaps

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kr/pretty"
)

var (
	testsRegistry = newRegistry()
	_m            = sync.Mutex{}
	// Matches [ Test... - number ] testIDs
	testIDRegexp         = regexp.MustCompile(`(?m)^\[(Test.* - \d)\]$`)
	endCharRegexp        = regexp.MustCompile(`(?m)(^---$)`)
	endCharEscapedRegexp = regexp.MustCompile(`(?m)(^/-/-/-/$)`)
)

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

func unescapeEndChars(input string) string {
	return endCharEscapedRegexp.ReplaceAllLiteralString(input, "---")
}

func escapeEndChars(input string) string {
	// This is for making sure a snapshot doesn't contain an ending char
	return endCharRegexp.ReplaceAllLiteralString(input, "/-/-/-/")
}

func getPrevSnapshot(testID, snapPath string) (string, error) {
	f, err := snapshotFileToString(snapPath)
	if err != nil {
		return "", err
	}

	match := dynamicTestIDRegexp(testID).FindStringSubmatch(f)

	if len(match) < 2 {
		return "", errSnapNotFound
	}

	// The second capture group contains the snapshot data
	return match[1], nil
}

func snapshotFileToString(name string) (string, error) {
	if _, err := os.Stat(name); err != nil {
		return "", errSnapNotFound
	}

	f, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}

	return string(f), nil
}

func stringToSnapshotFile(snap, name string) error {
	return os.WriteFile(name, []byte(snap), os.ModePerm)
}

func addNewSnapshot(testID, snapshot, dir, snapPath string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(snapPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n%s\n%s---\n", testID, snapshot)
	if err != nil {
		return err
	}

	return nil
}

/*
Returns the dir for snapshots [where the tests run + /snapsDirName]
and the name [dir + /snapsDirName + /<test-name>.snapsExtName]
*/
func snapDirAndName() (dir, name string) {
	callerPath := baseCaller(2)
	base := filepath.Base(callerPath)

	dir = filepath.Join(filepath.Dir(callerPath), snapsDir)
	name = filepath.Join(dir, strings.TrimSuffix(base, filepath.Ext(base))+snapsExt)

	return
}

func updateSnapshot(testID, snapshot, snapPath string) error {
	// When t.Parallel a test can override another snapshot as we dump
	// all snapshots
	_m.Lock()
	defer _m.Unlock()
	f, err := snapshotFileToString(snapPath)
	if err != nil {
		return err
	}

	updatedSnapFile := dynamicTestIDRegexp(testID).
		ReplaceAllLiteralString(f, fmt.Sprintf("%s\n%s---", testID, snapshot))

	return stringToSnapshotFile(updatedSnapFile, snapPath)
}
