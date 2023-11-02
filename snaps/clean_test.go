package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

// loadMockSnap loads a mock snap from the testdata directory
func loadMockSnap(t *testing.T, name string) []byte {
	t.Helper()
	snap, err := os.ReadFile(fmt.Sprintf("testdata/%s", name))
	if err != nil {
		t.Fatal(err)
	}

	return snap
}

func setupTempExamineFiles(
	t *testing.T,
	mockSnap1, mockSnap2 []byte,
) (map[string]map[string]int, string, string) {
	t.Helper()
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	files := []struct {
		name string
		data []byte
	}{
		{
			name: filepath.FromSlash(dir1 + "/test1.snap"),
			data: mockSnap1,
		},
		{
			name: filepath.FromSlash(dir2 + "/test2.snap"),
			data: mockSnap2,
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

	tests := map[string]map[string]int{
		files[0].name: {
			"TestDir1_1/TestSimple": 1,
			"TestDir1_2/TestSimple": 1,
			"TestDir1_3/TestSimple": 2,
			"TestCat":               1,
			"TestAlpha":             2,
			"TestBeta":              1,
		},
		files[1].name: {
			"TestDir2_1/TestSimple": 3,
			"TestDir2_2/TestSimple": 1,
			"TestCat":               1,
			"TestAlpha":             2,
			"TestBeta":              1,
		},
	}

	return tests, dir1, dir2
}

func TestExamineFiles(t *testing.T) {
	t.Run("should parse files", func(t *testing.T) {
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1"),
			loadMockSnap(t, "mock-snap-2"),
		)
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
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1"),
			loadMockSnap(t, "mock-snap-2"),
		)
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
		shouldUpdate, sort := false, false
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1"),
			loadMockSnap(t, "mock-snap-2"),
		)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate, sort)

		test.Equal(t, []string{}, obsolete)
		test.NoError(t, err)
	})

	t.Run("should report two obsolete snapshots and not change content", func(t *testing.T) {
		shouldUpdate, sort := false, false
		mockSnap1 := loadMockSnap(t, "mock-snap-1")
		mockSnap2 := loadMockSnap(t, "mock-snap-2")
		tests, dir1, dir2 := setupTempExamineFiles(t, mockSnap1, mockSnap2)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		// Reducing test occurrence to 1 meaning the second test was removed ( testid - 2 )
		tests[used[0]]["TestDir1_3/TestSimple"] = 1
		// Removing the test entirely
		delete(tests[used[1]], "TestDir2_2/TestSimple")

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate, sort)
		content1 := test.GetFileContent(t, used[0])
		content2 := test.GetFileContent(t, used[1])

		test.Equal(t, []string{"TestDir1_3/TestSimple - 2", "TestDir2_2/TestSimple - 1"}, obsolete)
		test.NoError(t, err)

		// Content of snaps is not changed
		test.Equal(t, mockSnap1, []byte(content1))
		test.Equal(t, mockSnap2, []byte(content2))
	})

	t.Run("should update the obsolete snap files", func(t *testing.T) {
		shouldUpdate, sort := true, false
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1"),
			loadMockSnap(t, "mock-snap-2"),
		)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		// removing tests from the map means those tests are no longer used
		delete(tests[used[0]], "TestDir1_3/TestSimple")
		delete(tests[used[1]], "TestDir2_1/TestSimple")

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate, sort)
		content1 := test.GetFileContent(t, used[0])
		content2 := test.GetFileContent(t, used[1])

		// !!unsorted
		expected1 := `
[TestDir1_2/TestSimple - 1]
int(10)
string hello world 1 2 1
---

[TestDir1_1/TestSimple - 1]

int(1)

string hello world 1 1 1

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
			"TestDir2_1/TestSimple - 3",
			"TestDir2_1/TestSimple - 2",
		},
			obsolete,
		)
		test.NoError(t, err)

		// Content of snaps is not changed
		test.Equal(t, expected1, content1)
		test.Equal(t, expected2, content2)
	})

	t.Run("should sort all tests", func(t *testing.T) {
		shouldUpdate, sort := false, true
		mockSnap1 := loadMockSnap(t, "mock-snap-sort-1")
		mockSnap2 := loadMockSnap(t, "mock-snap-sort-2")
		expectedMockSnap1 := loadMockSnap(t, "mock-snap-sort-1-sorted")
		expectedMockSnap2 := loadMockSnap(t, "mock-snap-sort-2-sorted")
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			mockSnap1,
			mockSnap2,
		)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		obsolete, err := examineSnaps(tests, used, "", shouldUpdate, sort)

		test.NoError(t, err)
		test.Equal(t, 0, len(obsolete))

		content1 := test.GetFileContent(t, filepath.FromSlash(dir1+"/test1.snap"))
		content2 := test.GetFileContent(t, filepath.FromSlash(dir2+"/test2.snap"))

		test.Equal(t, string(expectedMockSnap1), content1)
		test.Equal(t, string(expectedMockSnap2), content2)
	})

	t.Run(
		"should not update file if snaps are already sorted and shouldUpdate=false",
		func(t *testing.T) {
			shouldUpdate, sort := false, true
			mockSnap1 := loadMockSnap(t, "mock-snap-sort-1-sorted")
			mockSnap2 := loadMockSnap(t, "mock-snap-sort-2-sorted")
			tests, dir1, dir2 := setupTempExamineFiles(
				t,
				mockSnap1,
				mockSnap2,
			)
			used := []string{
				filepath.FromSlash(dir1 + "/test1.snap"),
				filepath.FromSlash(dir2 + "/test2.snap"),
			}

			// removing tests from the map means those tests are no longer used
			delete(tests[used[0]], "TestDir1_3/TestSimple")
			delete(tests[used[1]], "TestDir2_1/TestSimple")

			obsolete, err := examineSnaps(tests, used, "", shouldUpdate, sort)

			test.NoError(t, err)
			test.Equal(t, []string{
				"TestDir1_3/TestSimple - 1",
				"TestDir1_3/TestSimple - 2",
				"TestDir2_1/TestSimple - 1",
				"TestDir2_1/TestSimple - 2",
				"TestDir2_1/TestSimple - 3",
			},
				obsolete,
			)

			content1 := test.GetFileContent(t, filepath.FromSlash(dir1+"/test1.snap"))
			content2 := test.GetFileContent(t, filepath.FromSlash(dir2+"/test2.snap"))

			test.Equal(t, string(mockSnap1), content1)
			test.Equal(t, string(mockSnap2), content2)
		},
	)
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

func TestGetTestID(t *testing.T) {
	testCases := []struct {
		input      string
		expectedID string
		valid      bool
	}{
		{"[Test123 - Some Test]", "Test123 - Some Test", true},
		{"", "", false},
		{"Invalid input", "", false},
		{"[Test - Missing Closing Bracket", "", false},
		{"[TesGetTestID- No Space]", "", false},
		{"[Test/something - 10]", "Test/something - 10", true},
		// // must have [
		{"Test something 10]", "", false},
		// must have Test at the start
		{"TesGetTestID -   ]", "", false},
		// must have dash between test name and number
		{"[Test something 10]", "", false},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			id, ok := getTestID([]byte(tc.input))

			test.Equal(t, tc.valid, ok)
			test.Equal(t, tc.expectedID, id)
		})
	}
}
