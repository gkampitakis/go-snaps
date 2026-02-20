package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
		{
			name: filepath.FromSlash(dir1 + "TestSomething_my_test_1.snap"),
			data: []byte{},
		},
		{
			name: filepath.FromSlash(dir1 + "TestSomething_my_test_2.snap"),
			data: []byte{},
		},
		{
			name: filepath.FromSlash(dir1 + "TestSomething_my_test_3.snap"),
			data: []byte{},
		},
		{
			name: filepath.FromSlash(dir2 + "TestAnotherThing_my_test_1.snap"),
			data: []byte{},
		},
		{
			name: filepath.FromSlash(dir2 + "TestAnotherThing_my_simple_test_1.snap"),
			data: []byte{},
		},
		{
			name: filepath.FromSlash(dir2 + "TestAnotherThing_my_simple_test_2.snap"),
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
		obsolete, used, isDirty := examineFiles(tests, set{
			dir1 + "TestSomething_my_test_1.snap":           struct{}{},
			dir2 + "TestAnotherThing_my_simple_test_1.snap": struct{}{},
		}, "", false)

		obsoleteExpected := []string{
			filepath.FromSlash(dir1 + "/obsolete1.snap"),
			filepath.FromSlash(dir2 + "/obsolete2.snap"),
			filepath.FromSlash(dir1 + "TestSomething_my_test_2.snap"),
			filepath.FromSlash(dir1 + "TestSomething_my_test_3.snap"),
			filepath.FromSlash(dir2 + "TestAnotherThing_my_test_1.snap"),
			filepath.FromSlash(dir2 + "TestAnotherThing_my_simple_test_2.snap"),
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
		test.True(t, isDirty)
	})

	t.Run("should remove outdated files", func(t *testing.T) {
		shouldUpdate := true
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1"),
			loadMockSnap(t, "mock-snap-2"),
		)
		examineFiles(tests, set{
			dir1 + "TestSomething_my_test_1.snap":           struct{}{},
			dir2 + "TestAnotherThing_my_simple_test_1.snap": struct{}{},
		}, "", shouldUpdate)

		for _, obsoleteFilename := range []string{
			dir1 + "/obsolete1.snap",
			dir2 + "/obsolete2.snap",
			dir1 + "TestSomething_my_test_2.snap",
			dir1 + "TestSomething_my_test_3.snap",
			dir2 + "TestAnotherThing_my_test_1.snap",
			dir2 + "TestAnotherThing_my_simple_test_2.snap",
		} {
			if _, err := os.Stat(filepath.FromSlash(obsoleteFilename)); !errors.Is(
				err,
				os.ErrNotExist,
			) {
				t.Errorf("obsolete file %s not removed", obsoleteFilename)
			}
		}
	})
}

func TestExamineSnaps(t *testing.T) {
	testIdLabelMappings := map[string]string{
		"TestDir1_3/TestSimple - 1": "TestDir1_3/TestSimple - 1",
		"TestDir1_2/TestSimple - 1": "TestDir1_2/TestSimple - 1",
		"TestDir1_3/TestSimple - 2": "TestDir1_3/TestSimple - 2",
		"TestDir1_1/TestSimple - 1": "TestDir1_1/TestSimple - 1",
		"TestDir2_2/TestSimple - 1": "TestDir2_2/TestSimple - 1",
		"TestDir2_1/TestSimple - 1": "TestDir2_1/TestSimple - 1",
		"TestDir2_1/TestSimple - 3": "TestDir2_1/TestSimple - 3",
		"TestDir2_1/TestSimple - 2": "TestDir2_1/TestSimple - 2",
		"TestCat - 1":               "TestCat - 1",
		"TestAlpha - 2":             "TestAlpha - 2",
		"TestBeta - 1":              "TestBeta - 1",
		"TestAlpha - 1":             "TestAlpha - 1",
	}

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

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)

		test.Equal(t, []string{}, obsolete)
		test.NoError(t, err)
		test.False(t, isDirty)
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

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)
		content1 := test.GetFileContent(t, used[0])
		content2 := test.GetFileContent(t, used[1])

		test.Equal(t, []string{"TestDir1_3/TestSimple - 2", "TestDir2_2/TestSimple - 1"}, obsolete)
		test.NoError(t, err)

		// Content of snaps is not changed
		test.Equal(t, mockSnap1, []byte(content1))
		test.Equal(t, mockSnap2, []byte(content2))

		// And thus we are dirty since the contents do need changing
		test.True(t, isDirty)
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

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)
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

		// Content of snaps have been updated
		test.Equal(t, expected1, content1)
		test.Equal(t, expected2, content2)

		// And thus we are not dirty
		test.False(t, isDirty)
	})

	t.Run("should sort all tests when allowed to update", func(t *testing.T) {
		shouldUpdate, sort := true, true
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

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)

		test.NoError(t, err)
		test.Equal(t, 0, len(obsolete))

		content1 := test.GetFileContent(t, filepath.FromSlash(dir1+"/test1.snap"))
		content2 := test.GetFileContent(t, filepath.FromSlash(dir2+"/test2.snap"))

		// Content of snaps are now sorted
		test.Equal(t, string(expectedMockSnap1), content1)
		test.Equal(t, string(expectedMockSnap2), content2)

		// And thus we are not dirty
		test.False(t, isDirty)
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

			obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)

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

			// Content of snaps is not changed
			test.Equal(t, string(mockSnap1), content1)
			test.Equal(t, string(mockSnap2), content2)

			// And thus we are dirty, since there are obsolete snapshots that should be removed
			test.True(t, isDirty)
		},
	)
}

func TestExamineSnaps_WithLabels(t *testing.T) {
	testIdLabelMappings := map[string]string{
		"TestDir1_3/TestSimple - 1": "TestDir1_3/TestSimple - 1",
		"TestDir1_2/TestSimple - 1": "TestDir1_2/TestSimple - 1 - my snapshot",
		"TestDir1_3/TestSimple - 2": "TestDir1_3/TestSimple - 2",
		"TestDir1_1/TestSimple - 1": "TestDir1_1/TestSimple - 1 - this is another snapshot",

		"TestDir2_2/TestSimple - 1": "TestDir2_2/TestSimple - 1",
		"TestDir2_1/TestSimple - 1": "TestDir2_1/TestSimple - 1 - stdout",
		"TestDir2_1/TestSimple - 3": "TestDir2_1/TestSimple - 3",
		"TestDir2_1/TestSimple - 2": "TestDir2_1/TestSimple - 2 - stderr",
	}

	t.Run("should report no obsolete snapshots", func(t *testing.T) {
		shouldUpdate, sort := false, false
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1-labeled"),
			loadMockSnap(t, "mock-snap-2-labeled"),
		)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)

		test.Equal(t, []string{}, obsolete)
		test.NoError(t, err)
		test.False(t, isDirty)
	})

	t.Run("should report two obsolete snapshots and not change content", func(t *testing.T) {
		shouldUpdate, sort := false, false
		mockSnap1 := loadMockSnap(t, "mock-snap-1-labeled-renamed")
		mockSnap2 := loadMockSnap(t, "mock-snap-2-labeled")
		tests, dir1, dir2 := setupTempExamineFiles(t, mockSnap1, mockSnap2)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		// Reducing test occurrence to 1 meaning the second test was removed ( testid - 2 )
		tests[used[0]]["TestDir1_3/TestSimple"] = 1
		// Removing the test entirely
		delete(tests[used[1]], "TestDir2_2/TestSimple")

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)
		content1 := test.GetFileContent(t, used[0])
		content2 := test.GetFileContent(t, used[1])

		test.Equal(t, []string{"TestDir1_2/TestSimple - 1", "TestDir1_3/TestSimple - 2", "TestDir2_2/TestSimple - 1"}, obsolete)
		test.NoError(t, err)

		// Content of snaps is not changed
		test.Equal(t, mockSnap1, []byte(content1))
		test.Equal(t, mockSnap2, []byte(content2))

		// And thus we are dirty since the contents do need changing
		test.True(t, isDirty)
	})

	t.Run("should update the obsolete snap files", func(t *testing.T) {
		shouldUpdate, sort := true, false
		tests, dir1, dir2 := setupTempExamineFiles(
			t,
			loadMockSnap(t, "mock-snap-1-labeled-renamed"),
			loadMockSnap(t, "mock-snap-2-labeled"),
		)
		used := []string{
			filepath.FromSlash(dir1 + "/test1.snap"),
			filepath.FromSlash(dir2 + "/test2.snap"),
		}

		// removing tests from the map means those tests are no longer used
		delete(tests[used[0]], "TestDir1_3/TestSimple")

		obsolete, isDirty, err := examineSnaps(tests, used, "", 1, shouldUpdate, sort, testIdLabelMappings)
		content1 := test.GetFileContent(t, used[0])
		content2 := test.GetFileContent(t, used[1])

		// !!unsorted
		expected1 := `
[TestDir1_2/TestSimple - 1 - my snapshot]
int(10)
string hello world 1 2 1
---

[TestDir1_1/TestSimple - 1 - this is another snapshot]

int(1)

string hello world 1 1 1

---
`
		expected2 := `
[TestDir2_2/TestSimple - 1]
int(1000)
string hello world 2 2 1
---

[TestDir2_1/TestSimple - 1 - stdout]
int(1)
string hello world 2 1 1
---

[TestDir2_1/TestSimple - 3]
int(100)
string hello world 2 1 3
---

[TestDir2_1/TestSimple - 2 - stderr]
int(10)
string hello world 2 1 2
---
`

		test.Equal(t, []string{
			"TestDir1_3/TestSimple - 1",
			"TestDir1_2/TestSimple - 1",
			"TestDir1_3/TestSimple - 2",
		},
			obsolete,
		)
		test.NoError(t, err)

		// Content of snaps have been updated
		test.Equal(t, expected1, content1)
		test.Equal(t, expected2, content2)

		// And thus we are not dirty
		test.False(t, isDirty)
	})
}

func TestOccurrences(t *testing.T) {
	t.Run("when count 1", func(t *testing.T) {
		tests := map[string]int{
			"add_%d":      3,
			"subtract_%d": 1,
			"divide_%d":   2,
		}

		expected := set{
			"add_%d - 1":      {},
			"add_%d - 2":      {},
			"add_%d - 3":      {},
			"subtract_%d - 1": {},
			"divide_%d - 1":   {},
			"divide_%d - 2":   {},
		}

		expectedStandalone := set{
			"add_1":      {},
			"add_2":      {},
			"add_3":      {},
			"subtract_1": {},
			"divide_1":   {},
			"divide_2":   {},
		}

		test.Equal(t, expected, occurrences(tests, 1, snapshotOccurrenceFMT))
		test.Equal(t, expectedStandalone, occurrences(tests, 1, standaloneOccurrenceFMT))
	})

	t.Run("when count 3", func(t *testing.T) {
		tests := map[string]int{
			"add_%d":      12,
			"subtract_%d": 3,
			"divide_%d":   9,
		}

		expected := set{
			"add_%d - 1":      {},
			"add_%d - 2":      {},
			"add_%d - 3":      {},
			"add_%d - 4":      {},
			"subtract_%d - 1": {},
			"divide_%d - 1":   {},
			"divide_%d - 2":   {},
			"divide_%d - 3":   {},
		}

		expectedStandalone := set{
			"add_1":      {},
			"add_2":      {},
			"add_3":      {},
			"add_4":      {},
			"subtract_1": {},
			"divide_1":   {},
			"divide_2":   {},
			"divide_3":   {},
		}

		test.Equal(t, expected, occurrences(tests, 3, snapshotOccurrenceFMT))
		test.Equal(t, expectedStandalone, occurrences(tests, 3, standaloneOccurrenceFMT))
	})
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
		{"[Test/something - 10]", "Test/something - 10", true},
		{"[Test/something - 10 - my label]", "Test/something - 10 - my label", true},
		{input: "[Test/something - 100231231dsada]", expectedID: "", valid: false},
		{input: "[Test/something - 100231231 ]", expectedID: "", valid: false},
		{input: "[Test/something -100231231 ]", expectedID: "", valid: false},
		{input: "[Test/something- 100231231]", expectedID: "", valid: false},
		{input: "[Test/something - a ]", expectedID: "", valid: false},
		{input: "[Test/something - 100231231dsada - my label]", expectedID: "", valid: false},
		// todo: decide if this should actually be considered valid
		//  (if not, we should probably always string.Trim labels)
		{input: "[Test/something - 100231231 - my label ]", expectedID: "Test/something - 100231231 - my label ", valid: true},
		{input: "[Test/something -100231231 - my label ]", expectedID: "", valid: false},
		{input: "[Test/something - 100231231 -my label]", expectedID: "", valid: false},
		{input: "[Test/something - 100231231-my label]", expectedID: "", valid: false},
		{input: "[Test/something - 100231231- my label]", expectedID: "", valid: false},
		{input: "[Test/something- 100231231 - my label]", expectedID: "", valid: false},
		{input: "[Test/something - a ]", expectedID: "", valid: false},
		{input: "[Test/something - a]", expectedID: "", valid: false},
		{input: "[Test/something - a - my label]", expectedID: "", valid: false},
		{input: "[Test/something - a - my label ]", expectedID: "", valid: false},
		{"[Test123 - Some Test]", "", false},
		{"[Test123 - 1 - Some Test]", "Test123 - 1 - Some Test", true},
		{"", "", false},
		{"Invalid input", "", false},
		{"[Test - Missing Closing Bracket", "", false},
		{"[TesGetTestID- No Space]", "", false},
		// must have [
		{"Test something 10]", "", false},
		// must have Test at the start
		{"TesGetTestID -   ]", "", false},
		// must have dash between test name and number
		{"[Test something 10]", "", false},
		{"[Test/something - not a number]", "", false},
		{"s", "", false},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			// make sure that the capacity of b is len(tc.input), this way
			// indexing beyond the capacity will cause test to panic
			b := make([]byte, 0, len(tc.input))
			b = append(b, []byte(tc.input)...)
			id, _, ok := getTestID(b)

			test.Equal(t, tc.valid, ok)
			test.Equal(t, tc.expectedID, id)
		})
	}
}

func TestNaturalSort(t *testing.T) {
	t.Run("should sort in descending order", func(t *testing.T) {
		items := []string{
			"[TestExample/Test_Case_1#74 - 1]",
			"[TestExample/Test_Case_2#01 - 1 - b]",
			"[TestExample/Test_Case_1#05 - 1]",
			"[TestExample/Test_Case_1#09 - 1]",
			"[TestExample/Test_Case_2#01 - 2 - a]",
			"[TestExample/Test_Case_1#74 - 1 - my label]",
			"[TestExample - 1]",
			"[TestExample/Test_Case_1#71 - 1]",
			"[TestExample/Test_Case_2#01 - 3 - c]",
			"[TestExample/Test_Case_1#100 - 1 - another label]",
			"[TestExample/Test_Case_1#7 - 1]",
		}
		expected := []string{
			"[TestExample - 1]",
			"[TestExample/Test_Case_1#05 - 1]",
			"[TestExample/Test_Case_1#7 - 1]",
			"[TestExample/Test_Case_1#09 - 1]",
			"[TestExample/Test_Case_1#71 - 1]",
			"[TestExample/Test_Case_1#74 - 1 - my label]",
			"[TestExample/Test_Case_1#74 - 1]",
			"[TestExample/Test_Case_1#100 - 1 - another label]",
			"[TestExample/Test_Case_2#01 - 1 - b]",
			"[TestExample/Test_Case_2#01 - 2 - a]",
			"[TestExample/Test_Case_2#01 - 3 - c]",
		}

		slices.SortFunc(items, naturalSort)

		test.Equal(t, expected, items)
	})
}
