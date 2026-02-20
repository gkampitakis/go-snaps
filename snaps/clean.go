package snaps

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/maruel/natural"
)

var testEvents = newTestEvents()

const (
	erred uint8 = iota
	added
	updated
	passed
)

type events struct {
	items map[uint8]int
	sync.Mutex
}

func (e *events) register(event uint8) {
	e.Lock()
	defer e.Unlock()
	e.items[event]++
}

func newTestEvents() *events {
	return &events{
		items: make(map[uint8]int),
		Mutex: sync.Mutex{},
	}
}

type CleanOpts struct {
	// If set to true, `go-snaps` will sort the snapshots
	Sort bool
}

// Clean runs checks for identifying obsolete snapshots and prints a Test Summary.
//
// Must be called in a TestMain
//
//	func TestMain(m *testing.M) {
//	 v := m.Run()
//
//	 // After all tests have run `go-snaps` can check for unused snapshots
//	 dirty, err := snaps.Clean(m)
//	 if err != nil {
//		 fmt.Println("Error cleaning snaps:", err)
//		 os.Exit(1)
//	 }
//	 if dirty {
//		 fmt.Println("Some snapshots were outdated.")
//		 os.Exit(1)
//	 }
//
//	 os.Exit(v)
//	}
//
// Clean also supports options for sorting the snapshots
//
//	func TestMain(m *testing.M) {
//	 v := m.Run()
//
//	 // After all tests have run `go-snaps` will sort snapshots
//	 dirty, err := snaps.Clean(m, snaps.CleanOpts{Sort: true})
//	 if err != nil {
//		 fmt.Println("Error cleaning snaps:", err)
//		 os.Exit(1)
//	 }
//	 if dirty {
//		 fmt.Println("Some snapshots were outdated.")
//		 os.Exit(1)
//	 }
//
//	 os.Exit(v)
//	}
func Clean(m *testing.M, opts ...CleanOpts) (bool, error) {
	var opt CleanOpts
	if len(opts) != 0 {
		opt = opts[0]
	}
	// This is just for making sure Clean is called from TestMain
	_ = m
	runOnly := flag.Lookup("test.run").Value.String()
	count, _ := strconv.Atoi(flag.Lookup("test.count").Value.String())
	registeredStandaloneTests := occurrences(
		standaloneTestsRegistry.cleanup,
		count,
		standaloneOccurrenceFMT,
	)

	obsoleteFiles, usedFiles, filesDirty := examineFiles(
		testsRegistry.cleanup,
		registeredStandaloneTests,
		runOnly,
		shouldClean,
	)
	obsoleteTests, snapsDirty, err := examineSnaps(
		testsRegistry.cleanup,
		usedFiles,
		runOnly,
		count,
		shouldClean,
		opt.Sort,
		testsRegistry.labeled,
	)
	if err != nil {
		return snapsDirty || filesDirty, err
	}

	if s := summary(
		obsoleteFiles,
		obsoleteTests,
		len(skippedTests.values),
		testEvents.items,
		shouldClean,
	); s != "" {
		fmt.Println(s)
	}

	return filesDirty || snapsDirty, nil
}

// getTestID will return the testID if the line is in the form of [Test... - number]
func getTestID(b []byte) (string, string, bool) {
	if len(b) == 0 {
		return "", "", false
	}

	// needs to start with [Test and end with ]
	if !bytes.HasPrefix(b, []byte("[Test")) || b[len(b)-1] != ']' {
		return "", "", false
	}

	// needs to contain at least one ' - ' seperator
	firstSeparator := bytes.Index(b, []byte(" - "))
	if firstSeparator == -1 {
		return "", "", false
	}

	// if there is a label, there will be a second seperator
	secondSeparator := bytes.LastIndex(b, []byte(" - "))

	if secondSeparator == -1 || secondSeparator == firstSeparator {
		secondSeparator = len(b) - 1
	}

	// needs to have a number after the first separator
	if !isNumber(b[firstSeparator+3 : secondSeparator]) {
		return "", "", false
	}

	return string(b[1 : len(b)-1]), string(b[1:secondSeparator]), true
}

func isNumber(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] < '0' || b[i] > '9' {
			return false
		}
	}

	return true
}

// examineFiles traverses all the directories where snap tests where executed and checks
// if "orphan" snap files exist (files containing `.snap` in their name).
//
// If they do they are marked as obsolete and they are either deleted if `shouldUpdate=true` or printed on the console.
func examineFiles(
	registry map[string]map[string]int,
	registeredStandaloneTests set,
	runOnly string,
	shouldUpdate bool,
) (obsolete, used []string, dirtyFiles bool) {
	uniqueDirs := set{}

	for snapPaths := range registry {
		uniqueDirs[filepath.Dir(snapPaths)] = struct{}{}
	}

	for snapPaths := range registeredStandaloneTests {
		uniqueDirs[filepath.Dir(snapPaths)] = struct{}{}
	}

	for dir := range uniqueDirs {
		dirContents, _ := os.ReadDir(dir)

		for _, content := range dirContents {
			// this is a sanity check shouldn't have dirs inside the snapshot dirs
			// and only delete any `.snap` files
			if content.IsDir() || !strings.Contains(content.Name(), snapsExt) {
				continue
			}

			snapPath := filepath.Join(dir, content.Name())
			if _, called := registry[snapPath]; called {
				used = append(used, snapPath)
				continue
			}

			// if it's a standalone snapshot we don't add it to used list
			// as we don't need it for the next step, to examine individual snaps inside the file
			// as it contains only one
			if registeredStandaloneTests.Has(snapPath) {
				continue
			}

			if isFileSkipped(dir, content.Name(), runOnly) {
				continue
			}

			obsolete = append(obsolete, snapPath)

			if !shouldUpdate {
				continue
			}

			if err := os.Remove(snapPath); err != nil {
				fmt.Println(err)
			}
		}
	}

	return obsolete, used, len(obsolete) > 0
}

func examineSnaps(
	registry map[string]map[string]int,
	used []string,
	runOnly string,
	count int,
	shouldUpdate,
	sort bool,
	testIdLabelMappings map[string]string,
) ([]string, bool, error) {
	obsoleteTests := []string{}
	tests := map[string]string{}
	data := bytes.Buffer{}
	testIDs := []string{}
	var isDirty bool

	for _, snapPath := range used {
		f, err := os.OpenFile(snapPath, os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, isDirty, err
		}

		var needsUpdating bool

		registeredTests := occurrences(registry[snapPath], count, snapshotOccurrenceFMT)
		s := snapshotScanner(f)

		for s.Scan() {
			b := s.Bytes()
			// Check if line is a test id
			testIDWithLabel, testIDWithoutLabel, match := getTestID(b)
			if !match {
				continue
			}
			testIDs = append(testIDs, testIDWithLabel)

			currentTestIdWithLabel, ok := testIdLabelMappings[testIDWithoutLabel]

			if ok && currentTestIdWithLabel != testIDWithLabel {
				obsoleteTests = append(obsoleteTests, testIDWithLabel)
				needsUpdating = true

				removeSnapshot(s)
				continue
			}

			if !registeredTests.Has(testIDWithoutLabel) && !testSkipped(testIDWithoutLabel, runOnly) {
				obsoleteTests = append(obsoleteTests, testIDWithoutLabel)
				needsUpdating = true

				removeSnapshot(s)
				continue
			}

			for s.Scan() {
				line := s.Bytes()

				if bytes.Equal(line, endSequenceByteSlice) {
					tests[testIDWithLabel] = data.String()

					data.Reset()
					break
				}

				data.Write(line)
				data.WriteByte('\n')
			}
		}

		if err := s.Err(); err != nil {
			return nil, isDirty, err
		}

		needsSorting := sort && !slices.IsSortedFunc(testIDs, naturalSort)

		// if we're not allowed to update anything, just capture if the snapshot
		// needs cleaning, and then continue to the next snapshot
		if !shouldUpdate {
			if needsUpdating || needsSorting {
				isDirty = true
			}

			f.Close()

			clear(tests)
			testIDs = testIDs[:0]
			data.Reset()

			continue
		}

		if needsSorting {
			// sort testIDs
			slices.SortFunc(testIDs, naturalSort)
		}

		if err := overwriteFile(f, nil); err != nil {
			return nil, isDirty, err
		}

		for _, id := range testIDs {
			test, ok := tests[id]
			if !ok {
				continue
			}

			fmt.Fprintf(f, "\n[%s]\n%s%s\n", id, test, endSequence)
		}
		f.Close()

		clear(tests)
		testIDs = testIDs[:0]
		data.Reset()
	}

	return obsoleteTests, isDirty, nil
}

func summary(
	obsoleteFiles, obsoleteTests []string, NOskippedTests int,
	testEvents map[uint8]int,
	shouldUpdate bool,
) string {
	if len(obsoleteFiles) == 0 &&
		len(obsoleteTests) == 0 &&
		len(testEvents) == 0 &&
		NOskippedTests == 0 {
		return ""
	}

	var s strings.Builder

	objectSummaryList := func(objects []string, name string) {
		subject := name
		action := "obsolete"
		color := colors.Yellow
		if len(objects) > 1 {
			subject = name + "s"
		}
		if shouldUpdate {
			action = "removed"
			color = colors.Green
		}

		colors.Fprint(
			&s,
			color,
			fmt.Sprintf("\n%s%d snapshot %s %s\n", arrowSymbol, len(objects), subject, action),
		)

		for _, object := range objects {
			colors.Fprint(
				&s,
				colors.Dim,
				fmt.Sprintf("  %s %s%s\n", enterSymbol, bulletSymbol, object),
			)
		}
	}

	fmt.Fprintf(&s, "\n%s\n\n", colors.Sprint(colors.BoldWhite, "Snapshot Summary"))

	printEvent(&s, colors.Green, successSymbol, "passed", testEvents[passed])
	printEvent(&s, colors.Red, errorSymbol, "failed", testEvents[erred])
	printEvent(&s, colors.Green, updateSymbol, "added", testEvents[added])
	printEvent(&s, colors.Green, updateSymbol, "updated", testEvents[updated])
	printEvent(&s, colors.Yellow, skipSymbol, "skipped", NOskippedTests)

	if len(obsoleteFiles) > 0 {
		objectSummaryList(obsoleteFiles, "file")
	}

	if len(obsoleteTests) > 0 {
		objectSummaryList(obsoleteTests, "test")
	}

	if !shouldUpdate && len(obsoleteFiles)+len(obsoleteTests) > 0 {
		it := "it"

		if len(obsoleteFiles)+len(obsoleteTests) > 1 {
			it = "them"
		}

		colors.Fprint(
			&s,
			colors.Dim,
			fmt.Sprintf(
				"\nTo remove %s, re-run tests with `UPDATE_SNAPS=clean go test ./...`\n",
				it,
			),
		)
	}

	return s.String()
}

func printEvent(w io.Writer, color, symbol, verb string, events int) {
	if events == 0 {
		return
	}
	subject := "snapshot"
	if events > 1 {
		subject += "s"
	}

	colors.Fprint(w, color, fmt.Sprintf("%s%v %s %s\n", symbol, events, subject, verb))
}

func standaloneOccurrenceFMT(s string, i int) string {
	return fmt.Sprintf(s, i)
}

func snapshotOccurrenceFMT(s string, i int) string {
	return fmt.Sprintf("%s - %d", s, i)
}

// Builds a Set with all snapshot ids registered. It uses the provider formatter to build keys.
func occurrences(tests map[string]int, count int, formatter func(string, int) string) set {
	result := make(set, len(tests))
	for testID, counter := range tests {
		// divide a test's counter by count (how many times the go test suite ran)
		// this gives us how many snapshots were created in a single test run.
		counter = counter / count
		if counter > 1 {
			for i := 1; i <= counter; i++ {
				result[formatter(testID, i)] = struct{}{}
			}
		}
		result[formatter(testID, counter)] = struct{}{}
	}

	return result
}

// naturalSort is a function that can be used to sort strings in natural order
func naturalSort(a, b string) int {
	if a == b {
		return 0
	}
	if natural.Less(a, b) {
		return -1
	}
	return 1
}
