package snaps

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func Equal(t *testing.T, expected interface{}, received interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, received) {
		t.Error(redText(fmt.Sprintf("[expected]: %v - [received]: %v", expected, received)))
	}
}

func createTempFile(t *testing.T, data string) string {
	path := t.TempDir()
	fPath := filepath.Join(path, "mock.file")

	_ = os.WriteFile(fPath, []byte(data), os.ModePerm)

	return fPath
}

func TestConfig(t *testing.T) {
	t.Run("New should return default config", func(t *testing.T) {
		c := New()

		Equal(t, "__snapshots__", c.snapsDir)
		Equal(t, "snap", c.snapsExt)
	})

	t.Run("should apply new options", func(t *testing.T) {
		c := New(SnapsDirectory("test_path"), SnapsExtension("mock_ext"))

		Equal(t, "test_path", c.snapsDir)
		Equal(t, "mock_ext", c.snapsExt)
	})
}

func TestConfigMethods(t *testing.T) {
	unitTestsPath := "mock-path"
	t.Cleanup(func() {
		err := os.RemoveAll(unitTestsPath)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("getPrevSnapshot", func(t *testing.T) {
		t.Run("should return errSnapNotFound", func(t *testing.T) {
			snap, err := defaultConfig.getPrevSnapshot("", "")

			Equal(t, "", snap)
			Equal(t, err, errSnapNotFound)
		})

		t.Run("should return errSnapNotFound if no match found", func(t *testing.T) {
			fileData := "test"
			path := createTempFile(t, fileData)

			snap, err := defaultConfig.getPrevSnapshot("", path)

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
				fileData: `[my-test - 1]
			mysnapshot
			---`,
				snap: "\t\t\tmysnapshot\n\t\t\t",
				err:  nil,
			},
			{
				description: "should ignore regex in testID and match correct snap",
				testID:      ".*",
				fileData: `
				[my-test]
				wrong snap
				---
	
				[.*]
				mysnapshot
				---
			`,
				snap: "\n\t\t\t\tmysnapshot\n\t\t\t\t",
				err:  nil,
			},
		} {
			s := scenario
			t.Run(s.description, func(t *testing.T) {
				t.Parallel()

				path := createTempFile(t, s.fileData)

				snap, err := defaultConfig.getPrevSnapshot(s.testID, path)

				Equal(t, s.err, err)
				Equal(t, s.snap, snap)
			})
		}
	})

	t.Run("addNewSnapshot", func(t *testing.T) {
		s := New(SnapsDirectory(unitTestsPath))

		p, f := s.snapPathAndFile()
		err := s.addNewSnapshot("[mock-id]", "my-snap\n", p, f)

		Equal(t, nil, err)

		snap, err := s.snapshotFileToString(f)

		Equal(t, nil, err)
		Equal(t, "\n[mock-id]\nmy-snap\n---\n", snap)
	})

	t.Run("snapPathAndFile", func(t *testing.T) {
		path, file := defaultConfig.snapPathAndFile()

		Equal(t, true, strings.Contains(path, "/snaps/__snapshots__"))
		Equal(t, true, strings.Contains(file, "/snaps/__snapshots__/config_test.snap"))
	})
}
