package snaps

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Clean() {
	if _, fName := baseCaller(); fName == "testing.tRunner" {
		fmt.Println(yellowText("go-snaps [Warning]: snaps.Clean should only run in 'TestMain'"))
		return
	}
	runOnly := parseRunFlag(os.Args)

	obsoleteSnaps, usedSnaps := parseFiles(testsOccur.values, runOnly, shouldUpdate)
	obsoleteTests, err := parseSnaps(testsOccur.values, usedSnaps, runOnly, shouldUpdate)
	if err != nil {
		fmt.Println(err)
	}

	summary(obsoleteSnaps, obsoleteTests)
}

/*
	Map containing the occurrences is checked against the filesystem.

	If a file exists but is not registered in the map we check if the file is
	skipped. (We do that by checking if the mod is imported and there is a call to
	`Matchsnapshot`). If not skipped and not registed means it's an obsolete snap file
	and we mark it as one.
*/
func parseFiles(testsOccur map[string]map[string]int, runOnly string, shouldUpdate bool) (obsolete []string, used []string) {
	uniqueDirs := map[string]struct{}{}

	for paths := range testsOccur {
		uniqueDirs[filepath.Dir(paths)] = struct{}{}
	}

	for dir := range uniqueDirs {
		dirContents, _ := ioutil.ReadDir(dir)

		for _, content := range dirContents {
			// this is a sanity check shouldn't have dirs inside the snapshot dirs
			if content.IsDir() {
				continue
			}

			file := filepath.Join(dir, content.Name())

			if _, called := testsOccur[file]; !called {
				isSkipped := isFileSkipped(dir, content.Name(), runOnly)
				if isSkipped {
					continue
				}

				if shouldUpdate {
					err := os.Remove(file)
					if err != nil {
						fmt.Println(err)
					}
				}

				obsolete = append(obsolete, file)
				continue
			}

			used = append(used, file)
		}
	}

	return obsolete, used
}

func parseSnaps(testsOccur map[string]map[string]int, used []string, runOnly string, shouldUpdate bool) ([]string, error) {
	obsoleteTests := []string{}

	for _, snapPaths := range used {
		hasDiffs := false
		updatedFile := ""

		f, err := os.Open(snapPaths)
		if err != nil {
			return nil, err
		}

		registeredTests := occurrences(testsOccur[snapPaths])
		s := bufio.NewScanner(f)

		for s.Scan() {
			text := s.Text()
			// Check if line is a test id
			match := testIDRegexp.FindStringSubmatch(text)
			if len(match) <= 1 {
				// we are skipping excessive empty lines
				if text != "" {
					updatedFile += text + "\n"
				}
				continue
			}
			testID := match[1]

			if _, exists := registeredTests[testID]; !exists && !testSkipped(testID, runOnly) {
				obsoleteTests = append(obsoleteTests, testID)
				hasDiffs = true

				for s.Scan() {
					// skip until ---
					if s.Text() == "---" {
						break
					}
				}

				continue
			}

			updatedFile += "\n" + text + "\n"
		}

		f.Close()
		if !hasDiffs || !shouldUpdate {
			continue
		}

		err = os.WriteFile(snapPaths, []byte(updatedFile), os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}

	return obsoleteTests, nil
}

func summary(obsoleteSnaps []string, obsoleteTests []string) {
	if len(obsoleteSnaps) == 0 && len(obsoleteTests) == 0 {
		return
	}

	fmt.Printf(" %s	\n", greenBG("Snapshot Summary"))

	if len(obsoleteSnaps) > 0 {
		if len(obsoleteSnaps) > 2 {
			fmt.Println(yellowText(fmt.Sprintf("%d obsolete files detected", len(obsoleteSnaps))))
		} else {
			fmt.Println(yellowText(fmt.Sprintf("%d obsolete file detected", len(obsoleteSnaps))))
		}

		for _, file := range obsoleteSnaps {
			fmt.Println(dimText("	" + file))
		}
	}

	if len(obsoleteTests) > 0 {
		fmt.Println(yellowText(fmt.Sprintf("%d obsolete tests detected", len(obsoleteTests))))
		for _, test := range obsoleteTests {
			fmt.Println(dimText("	" + test))
		}
	}
}

/*
	Builds a Set with all snapshot ids registered inside a snap file
	Form: testname - number id

	tests have the form map[filepath]: map[testname]: <number of snapshots>
	e.g ./path/__snapshots__/add_test.snap map[TestAdd] 3

		will result to

		TestAdd - 1
		TestAdd - 2
		TestAdd - 3

	as it means there are 3 snapshots registered with that test
*/
func occurrences(tests map[string]int) Set {
	result := make(Set)
	for testID, counter := range tests {
		if counter > 1 {
			for i := 1; i <= counter; i++ {
				result[fmt.Sprintf("%s - %d", testID, i)] = struct{}{}
			}
		}
		result[fmt.Sprintf("%s - %d", testID, counter)] = struct{}{}
	}

	return result
}

func parseRunFlag(args []string) (runOnly string) {
	flag := "-test.run="

	for _, arg := range args {
		if strings.HasPrefix(arg, flag) {
			return strings.Split(arg, flag)[1]
		}
	}

	return ""
}
