package snaps

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
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

func setupTempExamineFiles(t *testing.T) (map[string]map[string]*testRuns, string, string) {
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
		{
			name: filepath.FromSlash(dir2 + "/should_not_delete.txt"),
			data: []byte{},
		},
	}

	for _, file := range files {
		err := os.WriteFile(file.name, file.data, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := map[string]map[string]*testRuns{
		files[0].name: {
			"TestDir1_1/TestSimple": &testRuns{
				times: 1,
			},
			"TestDir1_2/TestSimple": &testRuns{times: 1},
			"TestDir1_3/TestSimple": &testRuns{times: 2},
		},
		files[1].name: {
			"TestDir2_1/TestSimple": &testRuns{times: 3},
			"TestDir2_2/TestSimple": &testRuns{times: 1},
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

		test.Equal(t, obsoleteExpected, obsolete)
		test.Equal(t, usedExpected, used)
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
	t.Run("should report no obsolete snapshots", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(t)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}
		shouldUpdate := false

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate)

		test.Equal(t, []string{}, obsolete)
		test.Nil(t, err)
	})

	t.Run("should report two obsolete snapshots and not change content", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(t)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}
		shouldUpdate := false

		// Reducing test occurrence to 1 meaning the second test was removed ( testid - 2 )
		tests[used[0]]["TestDir1_3/TestSimple"].times = 1
		// Removing the test entirely
		delete(tests[used[1]], "TestDir2_2/TestSimple")

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate)
		content1 := getFileContent(t, used[0])
		content2 := getFileContent(t, used[1])

		test.Equal(t, []string{"TestDir1_3/TestSimple - 2", "TestDir2_2/TestSimple - 1"}, obsolete)
		test.Nil(t, err)

		// Content of snaps is not changed
		test.Equal(t, mockSnap1, content1)
		test.Equal(t, mockSnap2, content2)
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

		test.Equal(t, []string{
			"TestDir1_3/TestSimple - 1",
			"TestDir1_3/TestSimple - 2",
			"TestDir2_1/TestSimple - 1",
			"TestDir2_1/TestSimple - 2",
			"TestDir2_1/TestSimple - 3",
		},
			obsolete,
		)
		test.Nil(t, err)

		// Content of snaps is not changed
		test.Equal(t, expected1, content1)
		test.Equal(t, expected2, content2)
	})
}

func TestOccurrences(t *testing.T) {
	tests := map[string]*testRuns{
		"TestArithmetic/should add":      {times: 3},
		"TestArithmetic/should subtract": {times: 1},
		"TestArithmetic/should divide":   {times: 2},
	}

	expected := set{
		"TestArithmetic/should add - 1":      {},
		"TestArithmetic/should add - 2":      {},
		"TestArithmetic/should add - 3":      {},
		"TestArithmetic/should subtract - 1": {},
		"TestArithmetic/should divide - 1":   {},
		"TestArithmetic/should divide - 2":   {},
	}

	test.Equal(t, expected, occurrences(tests))
}

func TestSummary(t *testing.T) {
	for _, v := range []struct {
		name     string
		snapshot string
	}{
		{
			name:     "should print obsolete file",
			snapshot: summary([]string{"test0.snap"}, nil, 0, nil, false),
		},
		{
			name: "should print obsolete tests",
			snapshot: summary(
				nil,
				[]string{"TestMock/should_pass - 1", "TestMock/should_pass - 2"},
				0,
				nil,
				false,
			),
		},
		{
			name:     "should print updated file",
			snapshot: summary([]string{"test0.snap"}, nil, 0, nil, true),
		},
		{
			name:     "should print updated test",
			snapshot: summary(nil, []string{"TestMock/should_pass - 1"}, 0, nil, true),
		},
		{
			name:     "should return empty string",
			snapshot: summary(nil, nil, 0, nil, false),
		},
		{
			name: "should print events",
			snapshot: summary(nil, nil, 0, map[uint8]int{
				added:   5,
				erred:   100,
				updated: 3,
				passed:  10,
			}, false),
		},
		{
			name:     "should print number of skipped tests",
			snapshot: summary(nil, nil, 1, nil, true),
		},
		{
			name: "should print all summary",
			snapshot: summary(
				[]string{"test0.snap"},
				[]string{"TestMock/should_pass - 1"},
				5,
				map[uint8]int{
					added:   5,
					erred:   100,
					updated: 3,
					passed:  10,
				},
				false,
			),
		},
	} {
		// capture v
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			MatchSnapshot(t, v.snapshot)
		})
	}
}

func TestTestIDRegex(t *testing.T) {
	for _, tc := range []struct {
		input   string
		id      string
		matched bool
	}{
		{
			input:   "[Test/something - 10]",
			id:      "Test/something - 10",
			matched: true,
		},
		{
			// must have [
			input:   "Test/something - 10]",
			matched: false,
		},
		{
			// must have Test at the start
			input:   "[Tes/something - 10]",
			matched: false,
		},
		{
			// must have dash between test name and number
			input:   "[Test something 10]",
			matched: false,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if tc.matched {
				test.Equal(t, tc.id, testIDRegexp.FindStringSubmatch(tc.input)[1])
				return
			}

			test.Equal(t, 0, len(testIDRegexp.FindStringSubmatch(tc.input)))
		})
	}
}
