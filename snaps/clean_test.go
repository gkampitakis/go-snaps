package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

const mockSnap1 = `

[TestDir1_1/TestSimple - 1]
int(1)
string hello world 1 1 1
---

[TestDir1_2/TestSimple - 1]
int(10)
string hello world 1 2 1
---

[TestDir1_3/TestSimple - 1]
int(100)
string hello world 1 3 1
---

[TestDir1_3/TestSimple - 2]
int(1000)
string hello world 1 3 2
---

`

const mockSnap2 = `

[TestDir2_1/TestSimple - 1]
int(1)
string hello world 2 1 1
---

[TestDir2_1/TestSimple - 2]
int(10)
string hello world 2 1 2
---

[TestDir2_1/TestSimple - 3]
int(100)
string hello world 2 1 3
---

[TestDir2_2/TestSimple - 1]
int(1000)
string hello world 2 2 1
---

`

func setupTempExamineFiles(t *testing.T) (map[string]map[string]int, string, string) {
	t.Helper()
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	files := []struct {
		name string
		data []byte
	}{
		{
			name: filepath.FromSlash(dir1 + "/test1.snap"),
			data: []byte(mockSnap1),
		},
		{
			name: filepath.FromSlash(dir2 + "/test2.snap"),
			data: []byte(mockSnap2),
		},
		{
			name: filepath.FromSlash(dir1 + "/obsolete1.snap"),
			data: []byte{},
		},
		{
			name: filepath.FromSlash(dir2 + "/obsolete2.snap"),
			data: []byte{},
		},
	}

	for _, file := range files {
		err := os.WriteFile(file.name, file.data, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := map[string]map[string]int{
		files[0].name: {
			"TestDir1_1/TestSimple": 1,
			"TestDir1_2/TestSimple": 1,
			"TestDir1_3/TestSimple": 2,
		},
		files[1].name: {
			"TestDir2_1/TestSimple": 3,
			"TestDir2_2/TestSimple": 1,
		},
	}

	return tests, dir1, dir2
}

func getFileContent(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile(name)
	if err != nil {
		t.Error(err)
	}

	return string(content)
}

func TestExamineFiles(t *testing.T) {
	t.Run("should parse files", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(t)
		obsolete, used := examineFiles(tests, "", false)

		obsoleteExpected := []string{
			filepath.FromSlash(dir1 + "/obsolete1.snap"),
			filepath.FromSlash(dir2 + "/obsolete2.snap"),
		}
		usedExpected := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		// Parse files uses maps so order of strings cannot be guaranteed
		sort.Strings(obsoleteExpected)
		sort.Strings(usedExpected)
		sort.Strings(obsolete)
		sort.Strings(used)

		Equal(t, obsoleteExpected, obsolete)
		Equal(t, usedExpected, used)
	})

	t.Run("should remove outdate files", func(t *testing.T) {
		shouldUpdate := true
		tests, dir1, dir2 := setupTempExamineFiles(t)
		examineFiles(tests, "", shouldUpdate)

		if _, err := os.Stat(filepath.FromSlash(dir1 + "/obsolete1.snap")); !errors.Is(
			err,
			os.ErrNotExist,
		) {
			t.Error("obsolete obsolete1.snap not removed")
		}

		if _, err := os.Stat(filepath.FromSlash(dir2 + "/obsolete2.snap")); !errors.Is(
			err,
			os.ErrNotExist,
		) {
			t.Error("obsolete obsolete2.snap not removed")
		}
	})
}

func TestExamineSnaps(t *testing.T) {
	t.Run("should report no obsolete tests", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(t)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}
		shouldUpdate := false

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate)

		Equal(t, []string{}, obsolete)
		Equal(t, err, nil)
	})

	t.Run("should report two obsolete tests and not change content", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(t)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}
		shouldUpdate := false

		// Reducing test occurrence to 1 meaning the second test was removed ( testid - 2 )
		tests[used[0]]["TestDir1_3/TestSimple"] = 1
		// Removing the test entirely
		delete(tests[used[1]], "TestDir2_2/TestSimple")

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate)
		content1 := getFileContent(t, used[0])
		content2 := getFileContent(t, used[1])

		Equal(t, []string{"TestDir1_3/TestSimple - 2", "TestDir2_2/TestSimple - 1"}, obsolete)
		Equal(t, err, nil)

		// Content of snaps is not changed
		Equal(t, mockSnap1, content1)
		Equal(t, mockSnap2, content2)
	})

	t.Run("should update the obsolete snap files", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(t)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}
		shouldUpdate := true

		delete(tests[used[0]], "TestDir1_3/TestSimple")
		delete(tests[used[1]], "TestDir2_1/TestSimple")

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate)
		content1 := getFileContent(t, used[0])
		content2 := getFileContent(t, used[1])

		expected1 := `
[TestDir1_1/TestSimple - 1]
int(1)
string hello world 1 1 1
---

[TestDir1_2/TestSimple - 1]
int(10)
string hello world 1 2 1
---
`
		expected2 := `
[TestDir2_2/TestSimple - 1]
int(1000)
string hello world 2 2 1
---
`

		Equal(t, []string{
			"TestDir1_3/TestSimple - 1",
			"TestDir1_3/TestSimple - 2",
			"TestDir2_1/TestSimple - 1",
			"TestDir2_1/TestSimple - 2",
			"TestDir2_1/TestSimple - 3",
		},
			obsolete,
		)
		Equal(t, err, nil)

		// Content of snaps is not changed
		Equal(t, expected1, content1)
		Equal(t, expected2, content2)
	})
}

func TestOccurrences(t *testing.T) {
	tests := map[string]int{
		"add":      3,
		"subtract": 1,
		"divide":   2,
	}

	expected := set{
		"add - 1":      {},
		"add - 2":      {},
		"add - 3":      {},
		"subtract - 1": {},
		"divide - 1":   {},
		"divide - 2":   {},
	}

	Equal(t, expected, occurrences(tests))
}

func TestParseRunFlag(t *testing.T) {
	t.Run("should return empty string", func(t *testing.T) {
		runOly := parseRunFlag([]string{"-test.flag=ignore"})

		Equal(t, "", runOly)
	})

	t.Run("should return -run value", func(t *testing.T) {
		runOly := parseRunFlag([]string{"-test.run=MyTest"})

		Equal(t, "MyTest", runOly)
	})
}

func TestSummary(t *testing.T) {
	t.Run("should not print anything", func(t *testing.T) {
		mockPrinter := func(format string, args ...interface{}) (int, error) {
			NotCalled(t)
			return 0, nil
		}

		summary(mockPrinter, nil, nil, false)
	})

	t.Run("should print obsolete files", func(t *testing.T) {
		expectedCalls := []func(format string, args ...interface{}){
			func(format string, args ...interface{}) {
				expectedFormat := "\n%s\n\n"
				Equal(t, expectedFormat, format)

				expectedArg := greenBG("Snapshot Summary")
				Equal(t, expectedArg, args[0])
			},
			func(format string, arg ...interface{}) {
				expected := yellowText(
					fmt.Sprintf("%s%d snapshot files obsolete.\n", arrowPoint, 2),
				)
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("  " + bulletPoint + "test0.snap\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("  " + bulletPoint + "test1.snap\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := "\n"
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("You can remove obsolete files and tests by running 'UPDATE_SNAPS=true go test ./...'\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				NotCalled(t)
			},
		}
		mockPrinter := func(format string, args ...interface{}) (int, error) {
			expectedCalls[0](format, args...)

			expectedCalls = expectedCalls[1:]
			return 0, nil
		}

		summary(mockPrinter, []string{"test0.snap", "test1.snap"}, nil, false)
	})

	t.Run("should print obsolete tests", func(t *testing.T) {
		expectedCalls := []func(format string, args ...interface{}){
			func(format string, args ...interface{}) {
				expectedFormat := "\n%s\n\n"
				Equal(t, expectedFormat, format)

				expectedArg := greenBG("Snapshot Summary")
				Equal(t, expectedArg, args[0])
			},
			func(format string, arg ...interface{}) {
				expected := yellowText(
					fmt.Sprintf("%s%d snapshot tests obsolete.\n", arrowPoint, 2),
				)
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("  " + bulletPoint + "TestMock/should_pass - 1\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("  " + bulletPoint + "TestMock/should_pass - 2\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := "\n"
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("You can remove obsolete files and tests by running 'UPDATE_SNAPS=true go test ./...'\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				NotCalled(t)
			},
		}
		mockPrinter := func(format string, args ...interface{}) (int, error) {
			expectedCalls[0](format, args...)

			expectedCalls = expectedCalls[1:]
			return 0, nil
		}

		summary(mockPrinter, nil, []string{"TestMock/should_pass - 1", "TestMock/should_pass - 2"}, false)
	})

	t.Run("should print updated file", func(t *testing.T) {
		expectedCalls := []func(format string, args ...interface{}){
			func(format string, args ...interface{}) {
				expectedFormat := "\n%s\n\n"
				Equal(t, expectedFormat, format)

				expectedArg := greenBG("Snapshot Summary")
				Equal(t, expectedArg, args[0])
			},
			func(format string, arg ...interface{}) {
				expected := greenText(
					fmt.Sprintf("%s%d snapshot file removed.\n", arrowPoint, 1),
				)
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("  " + bulletPoint + "test0.snap\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := "\n"
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				NotCalled(t)
			},
		}
		mockPrinter := func(format string, args ...interface{}) (int, error) {
			expectedCalls[0](format, args...)

			expectedCalls = expectedCalls[1:]
			return 0, nil
		}

		summary(mockPrinter, []string{"test0.snap"}, nil, true)
	})

	t.Run("should print updated test", func(t *testing.T) {
		expectedCalls := []func(format string, args ...interface{}){
			func(format string, args ...interface{}) {
				expectedFormat := "\n%s\n\n"
				Equal(t, expectedFormat, format)

				expectedArg := greenBG("Snapshot Summary")
				Equal(t, expectedArg, args[0])
			},
			func(format string, arg ...interface{}) {
				expected := greenText(
					fmt.Sprintf("%s%d snapshot test removed.\n", arrowPoint, 1),
				)
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := dimText("  " + bulletPoint + "TestMock/should_pass - 1\n")
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				expected := "\n"
				Equal(t, expected, format)
			},
			func(format string, args ...interface{}) {
				NotCalled(t)
			},
		}
		mockPrinter := func(format string, args ...interface{}) (int, error) {
			expectedCalls[0](format, args...)

			expectedCalls = expectedCalls[1:]
			return 0, nil
		}

		summary(mockPrinter, nil, []string{"TestMock/should_pass - 1"}, true)
	})
}
