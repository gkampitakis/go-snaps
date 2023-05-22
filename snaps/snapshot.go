package snaps

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gkampitakis/go-snaps/internal/colors"
)

var (
	testsRegistry = newRegistry()
	_m            = sync.Mutex{}
)

var (
	addedMsg   = colors.Sprint(colors.Green, updateSymbol+"Snapshot added")
	updatedMsg = colors.Sprint(colors.Green, updateSymbol+"Snapshot updated")
)

type config struct {
	filename string
	snapsDir string
}

// Specify folder name where snapshots are stored
//
//	default: __snapshots__
//
// this doesn't change the file extension
func Filename(name string) func(*config) {
	return func(c *config) {
		c.filename = name
	}
}

// Specify folder name where snapshots are stored
//
//	default: __snapshots__
//
// Accepts absolute paths
func Dir(dir string) func(*config) {
	return func(c *config) {
		c.snapsDir = dir
	}
}

// Create snaps with configuration
//
//	e.g WithConfig(Filename("my_test")).MatchSnapshot(t, "hello world")
func WithConfig(args ...func(*config)) *config {
	s := defaultConfig

	for _, arg := range args {
		arg(&s)
	}

	return &s
}

func handleError(t testingT, err interface{}) {
	t.Helper()
	t.Error(err)
	testEvents.register(erred)
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

func newRegistry() *syncRegistry {
	return &syncRegistry{
		values: make(map[string]map[string]int),
		Mutex:  sync.Mutex{},
	}
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

/*
Returns the dir for snapshots
  - if no config provided returns the directory where tests are running
  - if snapsDir is relative path just gets appended to directory where tests are running
  - if snapsDir is absolute path then we are returning this path

Returns the filename
  - if no config provided we use the test file name with `.snap` extension
  - if filename provided we return the filename with `.snap` extension
*/
func snapDirAndName(c *config) (string, string) {
	//  skips current func, the wrapper match* and the exported Match* func
	callerPath := baseCaller(3)

	dir := c.snapsDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(filepath.Dir(callerPath), c.snapsDir)
	}

	filename := c.filename
	if filename == "" {
		base := filepath.Base(callerPath)
		filename = strings.TrimSuffix(base, filepath.Ext(base))
	}

	return dir, filepath.Join(dir, filename+snapsExt)
}

// Matches a specific testID
func dynamicTestIDRegexp(testID string) *regexp.Regexp {
	// e.g (?m)(?:\[TestAdd\/Hello_World\/my-test - 1\][\s\S])(.*[\s\S]*?)(?:^---$)
	return regexp.MustCompile(`(?m)(?:` + regexp.QuoteMeta(testID) + `[\s\S])(.*[\s\S]*?)(?:^---$)`)
}

func unescapeEndChars(s string) string {
	ss := strings.Split(s, "\n")
	for idx, s := range ss {
		if s == "/-/-/-/" {
			ss[idx] = "---"
		}
	}
	return strings.Join(ss, "\n")
}

func escapeEndChars(s string) string {
	ss := strings.Split(s, "\n")
	for idx, s := range ss {
		if s == "---" {
			ss[idx] = "/-/-/-/"
		}
	}
	return strings.Join(ss, "\n")
}
