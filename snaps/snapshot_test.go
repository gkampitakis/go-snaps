package snaps

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestTestID(t *testing.T) {
	t.Run("should increment id on each call [concurrent safe]", func(t *testing.T) {
		wg := sync.WaitGroup{}
		registry := newRegistry()

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				registry.getTestID("test", "/file")
				wg.Done()
			}()
		}

		wg.Wait()

		test.Equal(t, "[test - 6]", registry.getTestID("test", "/file"))
		test.Equal(t, "[test-v2 - 1]", registry.getTestID("test-v2", "/file"))
	})
}

func TestDynamicTestIDRegexp(t *testing.T) {
	t.Run("should escape regular expressions from testID", func(t *testing.T) {
		escapedTestIDRegexp := dynamicTestIDRegexp(`^\s+$-test`)
		expectedRegexp := `(?m)(?:\^\\s\+\$-test[\s\S])(.*[\s\S]*?)(?:^---$)`

		test.Equal(t, expectedRegexp, escapedTestIDRegexp.String())
	})
}

func TestGetPrevSnapshot(t *testing.T) {
	t.Run("should return errSnapNotFound", func(t *testing.T) {
		snap, err := getPrevSnapshot("", "")

		test.Equal(t, "", snap)
		test.Equal(t, err, errSnapNotFound)
	})

	t.Run("should return errSnapNotFound if no match found", func(t *testing.T) {
		fileData := "[testid]\ntest\n---\n"
		path := test.CreateTempFile(t, fileData)
		snap, err := getPrevSnapshot("nonexistentid", path)

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
			fileData:    "",
			snap:        "",
			err:         errSnapNotFound,
		},
		{
			description: "should not match",
			testID:      "my-test",
			fileData:    "mysnapshot",
			snap:        "",
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

			path := test.CreateTempFile(t, s.fileData)
			snap, err := getPrevSnapshot(s.testID, path)

			test.Equal(t, s.err, err)
			test.Equal(t, s.snap, snap)
		})
	}
}

func TestAddNewSnapshot(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "__snapshots__")
	name := filepath.Join(dir, "mock-test.snap")
	err := addNewSnapshot("[mock-id]", "my-snap\n", dir, name)

	test.Nil(t, err)

	snap, err := snapshotFileToString(name)

	test.Nil(t, err)
	test.Equal(t, "\n[mock-id]\nmy-snap\n---\n", snap)
}

func TestSnapPathAndFile(t *testing.T) {
	t.Run("should return default path and file", func(t *testing.T) {
		var (
			dir  string
			name string
		)

		func() {
			// This is for emulating being called from a func so we can find the correct file
			// of the caller
			func() {
				dir, name = snapDirAndName(&defaultConfig)
			}()
		}()

		test.Contains(t, dir, filepath.FromSlash("/snaps/__snapshots__"))
		test.Contains(t, name, filepath.FromSlash("/snaps/__snapshots__/snapshot_test.snap"))
	})

	t.Run("should return path and file from config", func(t *testing.T) {
		var (
			dir  string
			name string
		)

		func() {
			// This is for emulating being called from a func so we can find the correct file
			// of the caller
			func() {
				dir, name = snapDirAndName(&config{
					filename: "my_file",
					snapsDir: "my_snapshot_dir",
				})
			}()
		}()

		// returns the current file's path /snaps/*
		test.Contains(t, dir, filepath.FromSlash("/snaps/my_snapshot_dir"))
		test.Contains(t, name, filepath.FromSlash("/snaps/my_snapshot_dir/my_file.snap"))
	})

	t.Run("should return absolute path", func(t *testing.T) {
		var (
			dir  string
			name string
		)

		func() {
			// This is for emulating being called from a func so we can find the correct file
			// of the caller
			func() {
				dir, name = snapDirAndName(&config{
					filename: "my_file",
					snapsDir: "/path_to/my_snapshot_dir",
				})
			}()
		}()

		test.Contains(t, dir, filepath.FromSlash("/path_to/my_snapshot_dir"))
		test.Contains(t, name, filepath.FromSlash("/path_to/my_snapshot_dir/my_file.snap"))
	})
}

func TestUpdateSnapshot(t *testing.T) {
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
	snapPath := test.CreateTempFile(t, mockSnap)
	newSnapshot := "int(1250)\nstring new value\n"
	err := updateSnapshot("[Test_3/TestSimple - 1]", newSnapshot, snapPath)
	snap, _ := os.ReadFile(snapPath)

	test.Nil(t, err)
	test.Equal(t, updatedSnap, string(snap))
}

func TestEscapeEndChars(t *testing.T) {
	t.Run("should escape end chars inside data", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		snapshot := takeSnapshot([]interface{}{"my-snap", "---"})
		err := addNewSnapshot("[mock-id]", snapshot, dir, name)

		test.Nil(t, err)
		snap, err := snapshotFileToString(name)
		test.Nil(t, err)
		test.Equal(t, "\n[mock-id]\nmy-snap\n/-/-/-/\n---\n", snap)
	})

	t.Run("should not escape --- if not end chars", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "__snapshots__")
		name := filepath.Join(dir, "mock-test.snap")
		snapshot := takeSnapshot([]interface{}{"my-snap---", "---"})
		err := addNewSnapshot("[mock-id]", snapshot, dir, name)

		test.Nil(t, err)
		snap, err := snapshotFileToString(name)
		test.Nil(t, err)
		test.Equal(t, "\n[mock-id]\nmy-snap---\n/-/-/-/\n---\n", snap)
	})
}
