package snaps

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gkampitakis/go-snaps/internal/colors"
)

var (
	testsRegistry           = newRegistry()
	standaloneTestsRegistry = newStandaloneRegistry()
	_m                      = sync.RWMutex{}
	endSequenceByteSlice    = []byte(endSequence)
)

var (
	addedMsg   = colors.Sprint(colors.Green, updateSymbol+"Snapshot added")
	updatedMsg = colors.Sprint(colors.Green, updateSymbol+"Snapshot updated")
)

type Config struct {
	filename  string
	snapsDir  string
	extension string
	update    *bool
}

// Update determines whether to update snapshots or not
//
// It respects if running on CI.
func Update(u bool) func(*Config) {
	return func(c *Config) {
		c.update = &u
	}
}

// Specify folder name where snapshots are stored
//
//	default: __snapshots__
//
// this doesn't change the file extension see `snap.Ext`
func Filename(name string) func(*Config) {
	return func(c *Config) {
		c.filename = name
	}
}

// Specify folder name where snapshots are stored
//
//	default: __snapshots__
//
// Accepts absolute paths
func Dir(dir string) func(*Config) {
	return func(c *Config) {
		c.snapsDir = dir
	}
}

// Specify file name extension
//
// default: .snap
//
// Note: even if you specify a different extension the file still contain .snap
// e.g. if you specify .txt the file will be .snap.txt
func Ext(ext string) func(*Config) {
	return func(c *Config) {
		c.extension = ext
	}
}

// Create snaps with configuration
//
//	e.g snaps.WithConfig(snaps.Filename("my_test")).MatchSnapshot(t, "hello world")
func WithConfig(args ...func(*Config)) *Config {
	s := defaultConfig

	for _, arg := range args {
		arg(&s)
	}

	return &s
}

func handleError(t testingT, err any) {
	t.Helper()
	t.Error(err)
	testEvents.register(erred)
}

// We track occurrence as in the same test we can run multiple snapshots
// This also helps with keeping track with obsolete snaps
// map[snap path]: map[testname]: <number of snapshots>
type syncRegistry struct {
	running map[string]map[string]int
	cleanup map[string]map[string]int
	sync.Mutex
}

// Returns the id of the test in the snapshot
// Form [<test-name> - <occurrence>]
func (s *syncRegistry) getTestID(snapPath, testName string) string {
	s.Lock()

	if _, exists := s.running[snapPath]; !exists {
		s.running[snapPath] = make(map[string]int)
		s.cleanup[snapPath] = make(map[string]int)
	}

	s.running[snapPath][testName]++
	s.cleanup[snapPath][testName]++
	c := s.running[snapPath][testName]
	s.Unlock()

	return fmt.Sprintf("[%s - %d]", testName, c)
}

// reset sets only the number of running registry for the given test to 0.
func (s *syncRegistry) reset(snapPath, testName string) {
	s.Lock()
	s.running[snapPath][testName] = 0
	s.Unlock()
}

func newRegistry() *syncRegistry {
	return &syncRegistry{
		running: make(map[string]map[string]int),
		cleanup: make(map[string]map[string]int),
		Mutex:   sync.Mutex{},
	}
}

type syncStandaloneRegistry struct {
	running map[string]int
	cleanup map[string]int
	sync.Mutex
}

func newStandaloneRegistry() *syncStandaloneRegistry {
	return &syncStandaloneRegistry{
		running: make(map[string]int),
		cleanup: make(map[string]int),
		Mutex:   sync.Mutex{},
	}
}

func (s *syncStandaloneRegistry) getTestID(snapPath, snapPathRel string) (string, string) {
	s.Lock()

	s.running[snapPath]++
	s.cleanup[snapPath]++
	c := s.running[snapPath]
	s.Unlock()

	return fmt.Sprintf(snapPath, c), fmt.Sprintf(snapPathRel, c)
}

func (s *syncStandaloneRegistry) reset(snapPath string) {
	s.Lock()
	s.running[snapPath] = 0
	s.Unlock()
}

// getPrevSnapshot scans file searching for a snapshot matching the given testID and returns
// the snapshot with the line where is located inside the file.
//
// If not found returns errSnapNotFound error.
func getPrevSnapshot(testID, snapPath string) (string, int, error) {
	_m.RLock()
	defer _m.RUnlock()

	f, err := os.ReadFile(snapPath)
	if err != nil {
		return "", -1, errSnapNotFound
	}

	lineNumber := 1
	tid := []byte(testID)

	s := snapshotScanner(bytes.NewReader(f))
	for s.Scan() {
		l := s.Bytes()
		if !bytes.Equal(l, tid) {
			lineNumber++
			continue
		}
		var snapshot strings.Builder

		for s.Scan() {
			line := s.Bytes()

			if bytes.Equal(line, endSequenceByteSlice) {
				return strings.TrimSuffix(snapshot.String(), "\n"), lineNumber, nil
			}
			snapshot.Write(line)
			snapshot.WriteByte('\n')
		}
	}

	if err := s.Err(); err != nil {
		return "", -1, err
	}

	return "", -1, errSnapNotFound
}

func addNewSnapshot(testID, snapshot, snapPath string) error {
	if err := os.MkdirAll(filepath.Dir(snapPath), os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(snapPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "\n%s\n%s\n---\n", testID, snapshot)
	return err
}

func updateSnapshot(testID, snapshot, snapPath string) error {
	// When t.Parallel a test can override another snapshot as we dump
	// all snapshots
	_m.Lock()
	defer _m.Unlock()
	f, err := os.OpenFile(snapPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	tid := []byte(testID)
	var updatedSnapFile bytes.Buffer
	i, err := f.Stat()
	if err == nil {
		updatedSnapFile.Grow(int(i.Size()))
	}

	s := snapshotScanner(f)
	for s.Scan() {
		b := s.Bytes()
		updatedSnapFile.Write(b)
		updatedSnapFile.WriteByte('\n')
		if !bytes.Equal(b, tid) {
			continue
		}

		removeSnapshot(s)

		// add new snapshot
		updatedSnapFile.WriteString(snapshot)
		updatedSnapFile.WriteByte('\n')
		updatedSnapFile.Write(endSequenceByteSlice)
		updatedSnapFile.WriteByte('\n')
	}

	if err := s.Err(); err != nil {
		return err
	}

	return overwriteFile(f, updatedSnapFile.Bytes())
}

func overwriteFile(f *os.File, b []byte) error {
	f.Truncate(0)
	f.Seek(0, io.SeekStart)
	_, err := f.Write(b)
	return err
}

func removeSnapshot(s *bufio.Scanner) {
	for s.Scan() {
		// skip until ---
		if bytes.Equal(s.Bytes(), endSequenceByteSlice) {
			break
		}
	}
}

func upsertStandaloneSnapshot(snapshot, snapPath string) error {
	if err := os.MkdirAll(filepath.Dir(snapPath), os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(snapPath, []byte(snapshot), os.ModePerm)
}

func getPrevStandaloneSnapshot(snapPath string) (string, error) {
	f, err := os.ReadFile(snapPath)
	if err != nil {
		return "", errSnapNotFound
	}

	return string(f), nil
}

// Returns the path for snapshots:
//   - if no config provided returns the directory where tests are running
//   - if snapsDir is relative path just gets appended to directory where tests are running
//   - if snapsDir is absolute path then we are returning this path
//
// and for the filename:
//   - if no config provided we use the test file name with `.snap` extension
//   - if filename provided we return the filename with `.snap` extension
//   - if extension provided we return the filename with `.snap` and the provided extension
//   - if it's standalone snapshot we also append an integer (_%d) in the filename (even before `.snap`)
//
// Returns the relative path of the caller and the snapshot path.
func snapshotPath(c *Config, tName string, isStandalone bool) (string, string) {
	//  skips current func, the wrapper match* and the exported Match* func
	callerFilename := baseCaller(3)

	dir := c.snapsDir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(filepath.Dir(callerFilename), c.snapsDir)
	}

	snapPath := filepath.Join(dir, constructFilename(c, callerFilename, tName, isStandalone))
	snapPathRel, _ := filepath.Rel(filepath.Dir(callerFilename), snapPath)

	return snapPath, snapPathRel
}

func constructFilename(c *Config, callerFilename, tName string, isStandalone bool) string {
	filename := c.filename
	if filename == "" {
		base := filepath.Base(callerFilename)
		filename = strings.TrimSuffix(base, filepath.Ext(base))

		if isStandalone {
			filename = strings.ReplaceAll(tName, "/", "_")
		}
	}

	if isStandalone {
		filename += "_%d"
	}
	filename += snapsExt + c.extension

	return filename
}

func unescapeEndChars(s string) string {
	ss := strings.Split(s, "\n")
	for idx, s := range ss {
		if s == "/-/-/-/" {
			ss[idx] = endSequence
		}
	}
	return strings.Join(ss, "\n")
}

func escapeEndChars(s string) string {
	ss := strings.Split(s, "\n")
	for idx, s := range ss {
		if s == endSequence {
			ss[idx] = "/-/-/-/"
		}
	}
	return strings.Join(ss, "\n")
}
