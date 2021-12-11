package snaps

import (
	"errors"
	"os"
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

func setupTempParseFiles(t *testing.T) (map[string]map[string]int, string, string) {
	t.Helper()
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	files := []struct {
		name string
		data []byte
	}{
		{
			name: dir1 + "/test1.snap",
			data: []byte(mockSnap1),
		},
		{
			name: dir2 + "/test2.snap",
			data: []byte(mockSnap2),
		},
		{
			name: dir1 + "/obsolete1.snap",
			data: []byte{},
		},
		{
			name: dir2 + "/obsolete2.snap",
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

func TestParseFiles(t *testing.T) {
	t.Run("should parse files", func(t *testing.T) {
		tests, dir1, dir2 := setupTempParseFiles(t)
		obsolete, used := parseFiles(tests, false)

		obsoleteExpected := []string{dir1 + "/obsolete1.snap", dir2 + "/obsolete2.snap"}
		usedExpected := []string{dir1 + "/test1.snap", dir2 + "/test2.snap"}

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
		tests, dir1, dir2 := setupTempParseFiles(t)
		parseFiles(tests, shouldUpdate)

		if _, err := os.Stat(dir1 + "/obsolete1.snap"); !errors.Is(err, os.ErrNotExist) {
			t.Error("obsolete obsolete1.snap not removed")
		}

		if _, err := os.Stat(dir2 + "/obsolete2.snap"); !errors.Is(err, os.ErrNotExist) {
			t.Error("obsolete obsolete2.snap not removed")
		}
	})
}

func TestOccurrences(t *testing.T) {
	tests := map[string]int{
		"add":      3,
		"subtract": 1,
		"divide":   2,
	}

	expected := Set{
		"add - 1":      {},
		"add - 2":      {},
		"add - 3":      {},
		"subtract - 1": {},
		"divide - 1":   {},
		"divide - 2":   {},
	}

	Equal(t, expected, occurrences(tests))
}

func TestParseSnaps(t *testing.T) {
	t.Run("should report no obsolete tests", func(t *testing.T) {
		tests, dir1, dir2 := setupTempParseFiles(t)
		used := []string{dir1 + "/test1.snap", dir2 + "/test2.snap"}
		shouldUpdate := false

		obsolete, err := parseSnaps(tests, used, shouldUpdate)

		Equal(t, []string{}, obsolete)
		Equal(t, err, nil)
	})

	t.Run("should report two obsolete tests and not change content", func(t *testing.T) {
		tests, dir1, dir2 := setupTempParseFiles(t)
		used := []string{dir1 + "/test1.snap", dir2 + "/test2.snap"}
		shouldUpdate := false

		// Reducing test occurrence to 1 meaning the second test was removed ( testid - 2 )
		tests[used[0]]["TestDir1_3/TestSimple"] = 1
		// Removing the test entirely
		delete(tests[used[1]], "TestDir2_2/TestSimple")

		obsolete, err := parseSnaps(tests, used, shouldUpdate)
		content1 := getFileContent(t, used[0])
		content2 := getFileContent(t, used[1])

		Equal(t, []string{"TestDir1_3/TestSimple - 2", "TestDir2_2/TestSimple - 1"}, obsolete)
		Equal(t, err, nil)

		// Content of snaps is not changed
		Equal(t, mockSnap1, content1)
		Equal(t, mockSnap2, content2)
	})

	t.Run("should update the obsolete snap files", func(t *testing.T) {
		tests, dir1, dir2 := setupTempParseFiles(t)
		used := []string{dir1 + "/test1.snap", dir2 + "/test2.snap"}
		shouldUpdate := true

		delete(tests[used[0]], "TestDir1_3/TestSimple")
		delete(tests[used[1]], "TestDir2_1/TestSimple")

		obsolete, err := parseSnaps(tests, used, shouldUpdate)
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
