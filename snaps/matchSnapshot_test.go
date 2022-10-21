package snaps

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/ciinfo"
	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/internal/test"
)

const (
	fileName = "matchSnapshot_test.snap"
	mockSnap = `

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
)

func setupSnapshot(t *testing.T, filename string, ci bool, update ...bool) string {
	t.Helper()
	dir, _ := os.Getwd()
	snapPath := filepath.Join(dir, "__snapshots__", filename)
	isCI = ci
	shouldUpdatePrev := shouldUpdate
	shouldUpdate = false
	if len(update) > 0 {
		shouldUpdate = update[0]
	}

	t.Cleanup(func() {
		os.Remove(snapPath)
		testsRegistry = newRegistry()
		testEvents = newTestEvents()
		isCI = ciinfo.IsCI
		shouldUpdate = shouldUpdatePrev
	})

	_, err := os.Stat(snapPath)
	// This is for checking we are starting with a clean state testing
	test.True(t, errors.Is(err, os.ErrNotExist))

	return snapPath
}

func TestMatchSnapshot(t *testing.T) {
	t.Run("should create snapshot", func(t *testing.T) {
		snapPath := setupSnapshot(t, fileName, false)

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				test.NotCalled(t)
			},
			MockLog: func(args ...interface{}) { test.Equal(t, addedMsg, args[0]) },
		}

		MatchSnapshot(mockT, 10, "hello world")

		snap, err := snapshotFileToString(snapPath)
		test.Nil(t, err)
		test.Equal(t, "\n[mock-name - 1]\nint(10)\nhello world\n---\n", snap)
		test.Equal(t, 1, testEvents.items[added])
	})

	t.Run("if it's running on ci should skip", func(t *testing.T) {
		snapPath := setupSnapshot(t, fileName, true)

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				test.Equal(t, errSnapNotFound, args[0])
			},
			MockLog: func(args ...interface{}) {
				test.NotCalled(t)
			},
		}

		MatchSnapshot(mockT, 10, "hello world")

		_, err := snapshotFileToString(snapPath)
		test.Equal(t, errSnapNotFound, err)
		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should return error when diff is found", func(t *testing.T) {
		setupSnapshot(t, fileName, false)
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) { test.Equal(t, addedMsg, received) },
			func(received interface{}) { test.NotCalled(t) },
		}

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 2\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ Received + 2\x1b[0m\n\n\x1b[38;5;52m\x1b[48;5;225m- int(10)\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m" +
					"- hello world\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ int(100)\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ bye world\x1b[0m\n  \x1b[2m\n\x1b[0m"

				test.Equal(t, expected, args[0])
			},
			MockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		MatchSnapshot(mockT, 10, "hello world")
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, 100, "bye world")
		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		snapPath := setupSnapshot(t, fileName, false, true)
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) { test.Equal(t, addedMsg, received) },
			func(received interface{}) { test.Equal(t, updatedMsg, received) },
		}

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				test.NotCalled(t)
			},
			MockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		MatchSnapshot(mockT, 10, "hello world")
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, 100, "bye world")

		snap, err := snapshotFileToString(snapPath)
		test.Nil(t, err)
		test.Equal(t, "\n[mock-name - 1]\nint(100)\nbye world\n---\n", snap)
		test.Equal(t, 1, testEvents.items[updated])
	})

	t.Run("should print warning if no params provided", func(t *testing.T) {
		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockLog: func(args ...interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Yellow, "[warning] MatchSnapshot call without params\n"),
					args[0],
				)
			},
		}

		MatchSnapshot(mockT)
	})

	t.Run("diff should not print the escaped characters", func(t *testing.T) {
		setupSnapshot(t, fileName, false)
		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) { test.Equal(t, addedMsg, received) },
			func(received interface{}) { test.NotCalled(t) },
		}

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 3\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ Received + 3\x1b[0m\n\n\x1b[38;5;52m\x1b[48;5;225m- int(10)\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m" +
					"- hello world----\x1b[0m\n\x1b[38;5;52m\x1b[48;5;225m- ---\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ int(100)\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ bye world----\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
					"+ --\x1b[0m\n  \x1b[2m\n\x1b[0m"

				test.Equal(t, expected, args[0])
			},
			MockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot ( adding ending chars inside the diff )
		MatchSnapshot(mockT, 10, "hello world----", "---")

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, 100, "bye world----", "--")
	})
}
