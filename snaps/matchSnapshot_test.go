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

func TestMatchSnapshot(t *testing.T) {
	t.Run("should create snapshot", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "matchSnapshot_test.snap")
		isCI = false

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)

		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				test.NotCalled(t)
			},
			MockLog: func(args ...interface{}) {
				test.Equal(
					t,
					colors.Sprint(colors.Green, arrowPoint+"New snapshot written.\n"),
					args[0],
				)
			},
		}

		MatchSnapshot(mockT, 10, "hello world")

		snap, err := snapshotFileToString(snapPath)
		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-name - 1]\nint(10)\nhello world\n---\n", snap)
	})

	t.Run("if it's running on ci should skip", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "matchSnapshot_test.snap")
		isCI = true

		t.Cleanup(func() {
			os.Remove(snapPath)
			testsRegistry = newRegistry()
			isCI = ciinfo.IsCI
		})

		_, err := os.Stat(snapPath)

		test.Equal(t, true, errors.Is(err, os.ErrNotExist))

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

		_, err = snapshotFileToString(snapPath)
		test.Equal(t, errSnapNotFound, err)
	})

	t.Run("should return error when diff is found", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "matchSnapshot_test.snap")
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

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, 100, "bye world")
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "matchSnapshot_test.snap")
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

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchSnapshot(mockT, 100, "bye world")

		snap, err := snapshotFileToString(snapPath)
		test.Equal(t, nil, err)
		test.Equal(t, "\n[mock-name - 1]\nint(100)\nbye world\n---\n", snap)
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
		dir, _ := os.Getwd()
		snapPath := filepath.Join(dir, "__snapshots__", "matchSnapshot_test.snap")
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
