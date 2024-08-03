package snaps

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestSyncRegistry(t *testing.T) {
	t.Run("should increment id on each call [concurrent safe]", func(t *testing.T) {
		wg := sync.WaitGroup{}
		registry := newRegistry()

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				registry.getTestID("/file", "test")
				wg.Done()
			}()
		}

		wg.Wait()

		test.Equal(t, "[test - 6]", registry.getTestID("/file", "test"))
		test.Equal(t, "[test-v2 - 1]", registry.getTestID("/file", "test-v2"))
		test.Equal(t, registry.cleanup, registry.running)
	})

	t.Run("should reset running registry", func(t *testing.T) {
		wg := sync.WaitGroup{}
		registry := newRegistry()

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				registry.getTestID("/file", "test")
				wg.Done()
			}()
		}

		wg.Wait()

		registry.reset("/file", "test")

		// running registry start from 0 again
		test.Equal(t, "[test - 1]", registry.getTestID("/file", "test"))
		// cleanup registry still has 101
		test.Equal(t, 101, registry.cleanup["/file"]["test"])
	})
}

func TestSyncStandaloneRegistry(t *testing.T) {
	t.Run("should increment id on each call [concurrent safe]", func(t *testing.T) {
		wg := sync.WaitGroup{}
		registry := newStandaloneRegistry()

		for i := 0; i < 5; i++ {
			wg.Add(1)

			go func() {
				registry.getTestID("/file/my_file_%d.snap", "./__snapshots__/my_file_%d.snap")
				wg.Done()
			}()
		}

		wg.Wait()

		snapPath, snapPathRel := registry.getTestID(
			"/file/my_file_%d.snap",
			"./__snapshots__/my_file_%d.snap",
		)

		test.Equal(t, "/file/my_file_6.snap", snapPath)
		test.Equal(t, "./__snapshots__/my_file_6.snap", snapPathRel)

		snapPath, snapPathRel = registry.getTestID(
			"/file/my_other_file_%d.snap",
			"./__snapshots__/my_other_file_%d.snap",
		)

		test.Equal(t, "/file/my_other_file_1.snap", snapPath)
		test.Equal(t, "./__snapshots__/my_other_file_1.snap", snapPathRel)
		test.Equal(t, registry.cleanup, registry.running)
	})

	t.Run("should reset running registry", func(t *testing.T) {
		wg := sync.WaitGroup{}
		registry := newStandaloneRegistry()

		for i := 0; i < 100; i++ {
			wg.Add(1)

			go func() {
				registry.getTestID("/file/my_file_%d.snap", "./__snapshots__/my_file_%d.snap")
				wg.Done()
			}()
		}

		wg.Wait()

		registry.reset("/file/my_file_%d.snap")

		snapPath, snapPathRel := registry.getTestID(
			"/file/my_file_%d.snap",
			"./__snapshots__/my_file_%d.snap",
		)

		// running registry start from 0 again
		test.Equal(t, "/file/my_file_1.snap", snapPath)
		test.Equal(t, "./__snapshots__/my_file_1.snap", snapPathRel)
		// cleanup registry still has 101
		test.Equal(t, 101, registry.cleanup["/file/my_file_%d.snap"])
	})
}

func TestGetPrevSnapshot(t *testing.T) {
	t.Run("should return errSnapNotFound", func(t *testing.T) {
		snap, line, err := getPrevSnapshot("", "")

		test.Equal(t, "", snap)
		test.Equal(t, -1, line)
		test.Equal(t, errSnapNotFound, err)
	})

	t.Run("should return errSnapNotFound if no match found", func(t *testing.T) {
		fileData := "[testid]\ntest\n---\n"
		path := test.CreateTempFile(t, fileData)
		snap, line, err := getPrevSnapshot("nonexistentid", path)

		test.Equal(t, "", snap)
		test.Equal(t, -1, line)
		test.Equal(t, errSnapNotFound, err)
	})

	for _, scenario := range []struct {
		description string
		testID      string
		fileData    string
		snap        string
		line        int
		err         error
	}{
		{
			description: "should not match if no data",
			testID:      "my-test",
			fileData:    "",
			snap:        "",
			line:        -1,
			err:         errSnapNotFound,
		},
		{
			description: "should not match",
			testID:      "my-test",
			fileData:    "mysnapshot",
			snap:        "",
			line:        -1,
			err:         errSnapNotFound,
		},
		{
			description: "should return match",
			testID:      "[my-test - 1]",
			fileData:    "[my-test - 1]\nmysnapshot\n---\n",
			snap:        "mysnapshot",
			line:        1,
		},
		{
			description: "should ignore regex in testID and match correct snap",
			testID:      "[.*]",
			fileData:    "\n[my-test]\nwrong snap\n---\n\n[.*]\nmysnapshot\n---\n",
			snap:        "mysnapshot",
			line:        6,
		},
		{
			description: "should ignore end chars (---) inside snapshot",
			testID:      "[mock-test 1]",
			fileData:    "\n[mock-test 1]\nmysnapshot\n---moredata\n---\n",
			snap:        "mysnapshot\n---moredata",
			line:        2,
		},
	} {
		s := scenario
		t.Run(s.description, func(t *testing.T) {
			t.Parallel()

			path := test.CreateTempFile(t, s.fileData)
			snap, line, err := getPrevSnapshot(s.testID, path)

			test.Equal(t, s.err, err)
			test.Equal(t, s.line, line)
			test.Equal(t, s.snap, snap)
		})
	}
}

func TestAddNewSnapshot(t *testing.T) {
	snapPath := filepath.Join(t.TempDir(), "__snapshots__/mock-test.snap")

	test.NoError(t, addNewSnapshot("[mock-id]", "my-snap", snapPath))
	test.Equal(t, "\n[mock-id]\nmy-snap\n---\n", test.GetFileContent(t, snapPath))
}

func TestSnapshotPath(t *testing.T) {
	snapshotPathWrapper := func(c *Config, tName string, isStandalone bool) (snapPath, snapPathRel string) {
		// This is for emulating being called from a func so we can find the correct file
		// of the caller
		func() {
			func() {
				snapPath, snapPathRel = snapshotPath(c, tName, isStandalone)
			}()
		}()

		return
	}

	t.Run("should return default path and file", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&defaultConfig, "", false)

		test.HasSuffix(t, snapPath, filepath.FromSlash("/snaps/__snapshots__/snapshot_test.snap"))
		test.Equal(t, filepath.FromSlash("__snapshots__/snapshot_test.snap"), snapPathRel)
	})

	t.Run("should return path and file from config", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&Config{
			filename: "my_file",
			snapsDir: "my_snapshot_dir",
		}, "", false)

		// returns the current file's path /snaps/*
		test.HasSuffix(t, snapPath, filepath.FromSlash("/snaps/my_snapshot_dir/my_file.snap"))
		test.Equal(t, filepath.FromSlash("my_snapshot_dir/my_file.snap"), snapPathRel)
	})

	t.Run("should return absolute path", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&Config{
			filename: "my_file",
			snapsDir: "/path_to/my_snapshot_dir",
		}, "", false)

		test.HasSuffix(t, snapPath, filepath.FromSlash("/path_to/my_snapshot_dir/my_file.snap"))
		// the depth depends on filesystem structure
		test.HasSuffix(
			t,
			snapPathRel,
			filepath.FromSlash("path_to/my_snapshot_dir/my_file.snap"),
		)
	})

	t.Run("should add extension to filename", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&Config{
			filename:  "my_file",
			snapsDir:  "my_snapshot_dir",
			extension: ".txt",
		}, "", false)

		test.HasSuffix(t, snapPath, filepath.FromSlash("/snaps/my_snapshot_dir/my_file.snap.txt"))
		test.Equal(t, filepath.FromSlash("my_snapshot_dir/my_file.snap.txt"), snapPathRel)
	})

	t.Run("should return standalone snapPath", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&defaultConfig, "my_test", true)

		test.HasSuffix(
			t,
			snapPath,
			filepath.FromSlash("/snaps/__snapshots__/my_test_%d.snap"),
		)
		test.Equal(
			t,
			filepath.FromSlash("__snapshots__/my_test_%d.snap"),
			snapPathRel,
		)
	})

	t.Run("should return standalone snapPath without '/'", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&defaultConfig, "TestFunction/my_test", true)

		test.HasSuffix(
			t,
			snapPath,
			filepath.FromSlash("/snaps/__snapshots__/TestFunction_my_test_%d.snap"),
		)
		test.Equal(
			t,
			filepath.FromSlash("__snapshots__/TestFunction_my_test_%d.snap"),
			snapPathRel,
		)
	})

	t.Run("should return standalone snapPath with overridden filename", func(t *testing.T) {
		snapPath, snapPathRel := snapshotPathWrapper(&Config{
			filename: "my_file",
			snapsDir: "my_snapshot_dir",
		}, "my_test", true)

		test.HasSuffix(t, snapPath, filepath.FromSlash("/snaps/my_snapshot_dir/my_file_%d.snap"))
		test.Equal(t, filepath.FromSlash("my_snapshot_dir/my_file_%d.snap"), snapPathRel)
	})

	t.Run(
		"should return standalone snapPath with overridden filename and extension",
		func(t *testing.T) {
			snapPath, snapPathRel := snapshotPathWrapper(&Config{
				filename:  "my_file",
				snapsDir:  "my_snapshot_dir",
				extension: ".txt",
			}, "my_test", true)

			test.HasSuffix(
				t,
				snapPath,
				filepath.FromSlash("/snaps/my_snapshot_dir/my_file_%d.snap.txt"),
			)
			test.Equal(t, filepath.FromSlash("my_snapshot_dir/my_file_%d.snap.txt"), snapPathRel)
		},
	)
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
	newSnapshot := "int(1250)\nstring new value"

	test.NoError(t, updateSnapshot("[Test_3/TestSimple - 1]", newSnapshot, snapPath))
	test.Equal(t, updatedSnap, test.GetFileContent(t, snapPath))
}

func TestEscapeEndChars(t *testing.T) {
	t.Run("should escape end chars inside data", func(t *testing.T) {
		snapPath := filepath.Join(t.TempDir(), "__snapshots__/mock-test.snap")
		snapshot := takeSnapshot([]any{"my-snap", endSequence})

		test.NoError(t, addNewSnapshot("[mock-id]", snapshot, snapPath))
		test.Equal(t, "\n[mock-id]\nmy-snap\n/-/-/-/\n---\n", test.GetFileContent(t, snapPath))
	})

	t.Run("should not escape --- if not end chars", func(t *testing.T) {
		snapPath := filepath.Join(t.TempDir(), "__snapshots__/mock-test.snap")
		snapshot := takeSnapshot([]any{"my-snap---", endSequence})

		test.NoError(t, addNewSnapshot("[mock-id]", snapshot, snapPath))
		test.Equal(t, "\n[mock-id]\nmy-snap---\n/-/-/-/\n---\n", test.GetFileContent(t, snapPath))
	})
}
