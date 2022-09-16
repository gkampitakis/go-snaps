package snaps

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/ciinfo"
	"github.com/gkampitakis/go-snaps/snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/snaps/internal/test"
)

const mockSnap = `

[Test_1/TestSimple - 1]
int(1)
string hello world 1 1 1
---

[Test_3/TestSimple - 1]
int(100)
string hello world 1 3 1
---

[Test_3/TestSimple - 2]
int(1000)
string hello world 1 3 2
---

`

// Testing Helper Functions - Start

func createTempFile(t *testing.T, data string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "mock.file")

	_ = os.WriteFile(path, []byte(data), os.ModePerm)

	return path
}

// Testing Helper Functions - End

func TestInternalMethods(t *testing.T) {
	t.Run("getPrevSnapshot", func(t *testing.T) {
		t.Run("should return errSnapNotFound", func(t *testing.T) {
			snap, err := getPrevSnapshot("", "")

			test.Equal(t, "", snap)
			test.Equal(t, err, errSnapNotFound)
		})

		t.Run("should return errSnapNotFound if no match found", func(t *testing.T) {
			fileData := "test"
			path := createTempFile(t, fileData)
			snap, err := getPrevSnapshot("", path)

			test.Equal(t, "", snap)
			test.Equal(t, errSnapNotFound, err)
		})

		for _, scenario := range []struct {
			description string
			testID      string
			fileData    string
			snap        string
			err         error
		}{
			{
				description: "should not match if no data",
				testID:      "my-test",
				fileData:    ``,
				snap:        ``,
				err:         errSnapNotFound,
			},
			{
				description: "should not match",
				testID:      "my-test",
				fileData:    `mysnapshot`,
				snap:        ``,
				err:         errSnapNotFound,
			},
			{
				description: "should return match",
				testID:      "[my-test - 1]",
				fileData:    "[my-test - 1]\nmysnapshot\n---\n",
				snap:        "mysnapshot\n",
				err:         nil,
			},
			{
				description: "should ignore regex in testID and match correct snap",
				testID:      "[.*]",
				fileData:    "\n[my-test]\nwrong snap\n---\n\n[.*]\nmysnapshot\n---\n",
				snap:        "mysnapshot\n",
				err:         nil,
			},
			{
				description: "should ignore end chars (---) inside snapshot",
				testID:      "[mock-test 1]",
				fileData:    "\n[mock-test 1]\nmysnapshot\n---moredata\n---\n",
				snap:        "mysnapshot\n---moredata\n",
				err:         nil,
			},
		} {
			s := scenario
			t.Run(s.description, func(t *testing.T) {
				t.Parallel()

				path := createTempFile(t, s.fileData)
				snap, err := getPrevSnapshot(s.testID, path)

				test.Equal(t, s.err, err)
				test.Equal(t, s.snap, snap)
			})
		}
	})

	t.Run("addNewSnapshot", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		err := addNewSnapshot("[mock-id]", "my-snap\n", dir, name)

		test.Equal(t, nil, err)

		snap, err := snapshotFileToString(name)

		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-id]\nmy-snap\n---\n", snap)
	})

	t.Run("snapPathAndFile", func(t *testing.T) {
		path, file := snapDirAndName()

		test.Contains(t, path, filepath.FromSlash("/snaps/__snapshots__"))
		test.Contains(t, file, filepath.FromSlash("/snaps/__snapshots__/snaps_test.snap"))
	})

	t.Run("updateSnapshot", func(t *testing.T) {
		const updatedSnap = `

[Test_1/TestSimple - 1]
int(1)
string hello world 1 1 1
---

[Test_3/TestSimple - 1]
int(1250)
string new value
---

[Test_3/TestSimple - 2]
int(1000)
string hello world 1 3 2
---

`
		snapPath := createTempFile(t, mockSnap)
		newSnapshot := "int(1250)\nstring new value\n"
		err := updateSnapshot("[Test_3/TestSimple - 1]", newSnapshot, snapPath)
		snap, _ := os.ReadFile(snapPath)

		test.Equal(t, nil, err)
		test.Equal(t, updatedSnap, string(snap))
	})
}

func TestMatchSnapshot(t *testing.T) {
	t.Run("should create snapshot", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		isCI = false

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)

		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				test.NotCalled(t)
			},
			mockLog: func(args ...interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Green, arrowPoint+"New snapshot written.\n"),
					args[0],
				)
			},
		}

		MatchSnapshot(mockT, "10\nhello world")

		snap, err := snapshotFileToString(snapPath)
		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-name - 1]\n10\nhello world\n---\n", snap)
	})

	t.Run("if it's running on ci should skip", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		isCI = true

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)

		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				test.Equal(t, errSnapNotFound, args[0])
			},
			mockLog: func(args ...interface{}) {
				test.NotCalled(t)
			},
		}

		MatchSnapshot(mockT, "10\nhello world")

		_, err = snapshotFileToString(snapPath)
		test.Equal(t, errSnapNotFound, err)
	})

	t.Run("should return error when diff is found", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Green, arrowPoint+"New snapshot written.\n"),
					received,
				)
			},
			func(received interface{}) { test.NotCalled(t) },
		}
		isCI = false

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)
		// This is for checking we are starting with a clean state testing
		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 2\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ Received + 2\x1b[0m\n\n\x1b[38;5;52m\x1b[48;5;225m- 10\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m" +
					"- hello world\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ 100\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ bye world\x1b[0m\n  \x1b[2m\n\x1b[0m"

				test.Equal(t, expected, args[0])
			},
			mockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		MatchSnapshot(mockT, "10\nhello world")

		// Resetting registry to emulate the same matchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, "100\nbye world")
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		isCI = false
		shouldUpdatePrev := shouldUpdate
		shouldUpdate = true
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Green, arrowPoint+"New snapshot written.\n"),
					received,
				)
			},
			func(received interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Green, arrowPoint+"Snapshot updated.\n"),
					received,
				)
			},
		}

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
			shouldUpdate = shouldUpdatePrev
		})

		_, err := os.Stat(snapPath)
		// This is for checking we are starting with a clean state testing
		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				test.NotCalled(t)
			},
			mockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		MatchSnapshot(mockT, "10\nhello world")

		// Resetting registry to emulate the same matchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, "100\nbye world")

		snap, err := snapshotFileToString(snapPath)
		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-name - 1]\n100\nbye world\n---\n", snap)
	})

	t.Run("diff should not print the escaped characters", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Green, arrowPoint+"New snapshot written.\n"),
					received,
				)
			},
			func(received interface{}) { test.NotCalled(t) },
		}
		isCI = false

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)
		// This is for checking we are starting with a clean state testing
		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 3\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ Received + 3\x1b[0m\n\n\x1b[38;5;52m\x1b[48;5;225m- 10\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m" +
					"- hello world----\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m- ---\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ 100\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ bye world----\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ --\x1b[0m\n  \x1b[2m\n\x1b[0m"

				test.Equal(t, expected, args[0])
			},
			mockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot ( adding ending chars inside the diff )
		MatchSnapshot(mockT, "10\nhello world----\n---")

		// Resetting registry to emulate the same matchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, "100\nbye world----\n--")
	})
}

func TestEscapeEndChars(t *testing.T) {
	t.Run("should escape end chars inside data", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		snapshot := takeSnapshot("my-snap\n---")
		err := addNewSnapshot("[mock-id]", snapshot, dir, name)

		test.Equal(t, nil, err)
		snap, err := snapshotFileToString(name)
		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-id]\nmy-snap\n/-/-/-/\n---\n", snap)
	})

	t.Run("should not escape --- if not end chars", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		snapshot := takeSnapshot("my-snap---\n---")
		err := addNewSnapshot("[mock-id]", snapshot, dir, name)

		test.Equal(t, nil, err)
		snap, err := snapshotFileToString(name)
		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-id]\nmy-snap---\n/-/-/-/\n---\n", snap)
	})
}
