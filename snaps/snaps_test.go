package snaps

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gkampitakis/ciinfo"
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

func Equal(t *testing.T, expected, received interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\n[expected]: %v\n[received]: %v", expected, received)
	}
}

func Contains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("\n [expected] %s to contain %s", s, substr)
	}
}

func NotCalled(t *testing.T) {
	t.Helper()

	t.Errorf("function was not expected to be called")
}

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

			Equal(t, "", snap)
			Equal(t, err, errSnapNotFound)
		})

		t.Run("should return errSnapNotFound if no match found", func(t *testing.T) {
			fileData := "test"
			path := createTempFile(t, fileData)
			snap, err := getPrevSnapshot("", path)

			Equal(t, "", snap)
			Equal(t, errSnapNotFound, err)
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

				Equal(t, s.err, err)
				Equal(t, s.snap, snap)
			})
		}
	})

	t.Run("addNewSnapshot", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		err := addNewSnapshot("[mock-id]", "my-snap\n", dir, name)

		Equal(t, nil, err)

		snap, err := snapshotFileToString(name)

		Equal(t, nil, err)
		Equal(t, "\n[mock-id]\nmy-snap\n---\n", snap)
	})

	t.Run("snapPathAndFile", func(t *testing.T) {
		path, file := snapDirAndName()

		Contains(t, path, filepath.FromSlash("/snaps/__snapshots__"))
		Contains(t, file, filepath.FromSlash("/snaps/__snapshots__/snaps_test.snap"))
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

		Equal(t, nil, err)
		Equal(t, updatedSnap, string(snap))
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

		Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				NotCalled(t)
			},
			mockLog: func(args ...interface{}) {
				Equal(t, sprintColored(green, arrowPoint+"New snapshot written.\n"), args[0])
			},
		}

		matchSnapshot(mockT, []interface{}{10, "hello world"})

		snap, err := snapshotFileToString(snapPath)
		Equal(t, nil, err)
		Equal(t, "\n[mock-name - 1]\nint(10)\nhello world\n---\n", snap)
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

		Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				Equal(t, errSnapNotFound, args[0])
			},
			mockLog: func(args ...interface{}) {
				NotCalled(t)
			},
		}

		matchSnapshot(mockT, []interface{}{10, "hello world"})

		_, err = snapshotFileToString(snapPath)
		Equal(t, errSnapNotFound, err)
	})

	t.Run("should return error when diff is found", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) {
				Equal(t, sprintColored(green, arrowPoint+"New snapshot written.\n"), received)
			},
			func(received interface{}) { NotCalled(t) },
		}
		isCI = false

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)
		// This is for checking we are starting with a clean state testing
		Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 2\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ Received + 2\x1b[0m\n\n\x1b[38;5;52m\x1b[48;5;225m- int(10)\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m" +
					"- hello world\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ int(100)\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ bye world\x1b[0m\n  \x1b[2m\n\x1b[0m"

				Equal(t, expected, args[0])
			},
			mockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		matchSnapshot(mockT, []interface{}{10, "hello world"})

		// Resetting registry to emulate the same matchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		matchSnapshot(mockT, []interface{}{100, "bye world"})
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		isCI = false
		shouldUpdatePrev := shouldUpdate
		shouldUpdate = true
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) {
				Equal(t, sprintColored(green, arrowPoint+"New snapshot written.\n"), received)
			},
			func(received interface{}) {
				Equal(t, sprintColored(green, arrowPoint+"Snapshot updated.\n"), received)
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
		Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				NotCalled(t)
			},
			mockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		matchSnapshot(mockT, []interface{}{10, "hello world"})

		// Resetting registry to emulate the same matchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		matchSnapshot(mockT, []interface{}{100, "bye world"})

		snap, err := snapshotFileToString(snapPath)
		Equal(t, nil, err)
		Equal(t, "\n[mock-name - 1]\nint(100)\nbye world\n---\n", snap)
	})

	t.Run("should print warning if no params provided", func(t *testing.T) {
		mockT := MockTestingT{
			mockHelper: func() {},
			mockLog: func(args ...interface{}) {
				Equal(t, sprintColored(yellow, "[warning] MatchSnapshot call without params\n"), args[0])
			},
		}

		matchSnapshot(mockT, []interface{}{})
		matchSnapshot(mockT, nil)
	})

	t.Run("diff should not print the escaped characters", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "snaps_test.snap")
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) {
				Equal(t, sprintColored(green, arrowPoint+"New snapshot written.\n"), received)
			},
			func(received interface{}) { NotCalled(t) },
		}
		isCI = false

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)
		// This is for checking we are starting with a clean state testing
		Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := MockTestingT{
			mockHelper: func() {},
			mockName: func() string {
				return "mock-name"
			},
			mockError: func(args ...interface{}) {
				expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 3\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ Received + 3\x1b[0m\n\n\x1b[38;5;52m\x1b[48;5;225m- int(10)\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m" +
					"- hello world----\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m- ---\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ int(100)\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ bye world----\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ --\x1b[0m\n  \x1b[2m\n\x1b[0m"

				Equal(t, expected, args[0])
			},
			mockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot ( adding ending chars inside the diff )
		matchSnapshot(mockT, []interface{}{10, "hello world----", "---"})

		// Resetting registry to emulate the same matchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		matchSnapshot(mockT, []interface{}{100, "bye world----", "--"})
	})
}

func TestEscapeEndChars(t *testing.T) {
	t.Run("should escape end chars inside data", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		snapshot := takeSnapshot([]interface{}{"my-snap", "---"})
		err := addNewSnapshot("[mock-id]", snapshot, dir, name)

		Equal(t, nil, err)
		snap, err := snapshotFileToString(name)
		Equal(t, nil, err)
		Equal(t, "\n[mock-id]\nmy-snap\n/-/-/-/\n---\n", snap)
	})

	t.Run("should not escape --- if not end chars", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		snapshot := takeSnapshot([]interface{}{"my-snap---", "---"})
		err := addNewSnapshot("[mock-id]", snapshot, dir, name)

		Equal(t, nil, err)
		snap, err := snapshotFileToString(name)
		Equal(t, nil, err)
		Equal(t, "\n[mock-id]\nmy-snap---\n/-/-/-/\n---\n", snap)
	})
}
