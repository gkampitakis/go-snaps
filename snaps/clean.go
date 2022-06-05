package snaps

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Clean runs checks for identifying obsolete snapshots and prints a Test Summary.
//
// Should be called in a TestMain
//  func TestMain(t *testing.M) {
//   v := t.Run()
//
//   // After all tests have run `go-snaps` can check for not used snapshots
//   snaps.Clean()
//
//   os.Exit(v)
//  }
func Clean() {
	if _, fName := baseCaller(); fName == "testing.tRunner" {
		fmt.Println(yellowText("go-snaps [Warning]: snaps.Clean should only run in 'TestMain'"))
		return
	}
	runOnly := flag.Lookup("test.run").Value.String()

	obsoleteFiles, usedFiles := examineFiles(testsRegistry.values, runOnly, shouldUpdate)
	obsoleteTests, err := examineSnaps(testsRegistry.values, usedFiles, runOnly, shouldUpdate)
	if err != nil {
		fmt.Println(err)
		return
	}

	summary(fmt.Printf, obsoleteFiles, obsoleteTests, shouldUpdate)
}

/*
	Map containing the occurrences is checked against the filesystem.

	If a file exists but is not registered in the map we check if the file is
	skipped. (We do that by checking if the mod is imported and there is a call to
	`MatchSnapshot`). If not skipped and not registered means it's an obsolete snap file
	and we mark it as one.
*/
func examineFiles(
	registry map[string]map[string]int,
	runOnly string,
	shouldUpdate bool,
) (obsolete []string, used []string) {
	uniqueDirs := set{}

	for snapPaths := range registry {
		uniqueDirs[filepath.Dir(snapPaths)] = struct{}{}
	}

	for dir := range uniqueDirs {
		dirContents, _ := ioutil.ReadDir(dir)

		for _, content := range dirContents {
			// this is a sanity check shouldn't have dirs inside the snapshot dirs
			if content.IsDir() {
				continue
			}

			snapPath := filepath.Join(dir, content.Name())
			if _, called := registry[snapPath]; called {
				used = append(used, snapPath)
				continue
			}

			if isFileSkipped(dir, content.Name(), runOnly) {
				continue
			}

			obsolete = append(obsolete, snapPath)

			if !shouldUpdate {
				continue
			}

			if err := os.Remove(snapPath); err != nil {
				fmt.Println(err)
			}
		}
	}

	return obsolete, used
}

func examineSnaps(
	registry map[string]map[string]int,
	used []string,
	runOnly string,
	shouldUpdate bool,
) ([]string, error) {
	removeSnapshot := func(s *bufio.Scanner) {
		for s.Scan() {
			// skip until ---
			if s.Text() == "---" {
				break
			}
		}
	}
	getSnapshot := func(s *bufio.Scanner) (snapshot string) {
		for s.Scan() {
			line := s.Text()
			// reached end char
			if s.Text() == "---" {
				break
			}
			snapshot += line + newLine
		}

		return snapshot
	}
	obsoleteTests := []string{}

	for _, snapPath := range used {
		hasDiffs := false
		updatedFile := ""

		f, err := os.Open(snapPath)
		if err != nil {
			return nil, err
		}

		registeredTests := occurrences(registry[snapPath])
		s := bufio.NewScanner(f)

		for s.Scan() {
			// Check if line is a test id
			match := testIDRegexp.FindStringSubmatch(s.Text())

			if len(match) <= 1 {
				continue
			}
			testID := match[1]

			if _, exists := registeredTests[testID]; !exists && !testSkipped(testID, runOnly) {
				obsoleteTests = append(obsoleteTests, testID)
				hasDiffs = true

				removeSnapshot(s)

				continue
			}

			updatedFile += fmt.Sprintf("\n[%s]\n%s---\n", testID, getSnapshot(s))
		}

		f.Close()
		if !hasDiffs || !shouldUpdate {
			continue
		}

		if err = stringToSnapshotFile(updatedFile, snapPath); err != nil {
			fmt.Println(err)
		}
	}

	return obsoleteTests, nil
}

func summary(print printerF, obsoleteFiles []string, obsoleteTests []string, shouldUpdate bool) {
	if len(obsoleteFiles) == 0 && len(obsoleteTests) == 0 {
		return
	}

	objectSummaryList := func(objects []string, name string) {
		plural := name + "s"
		print(summaryMsg(
			len(objects),
			stringTernary(len(objects) > 1, plural, name),
			shouldUpdate),
		)

		for _, object := range objects {
			print(dimText(fmt.Sprintf("  %s%s\n", bulletPoint, object)))
		}

		print(newLine)
	}

	print("\n%s\n\n", greenBG("Snapshot Summary"))

	if len(obsoleteFiles) > 0 {
		objectSummaryList(obsoleteFiles, "file")
	}

	if len(obsoleteTests) > 0 {
		objectSummaryList(obsoleteTests, "test")
	}

	if !shouldUpdate {
		print(
			dimText(
				"You can remove obsolete files and snapshots by running 'UPDATE_SNAPS=true go test ./...'\n",
			),
		)
	}
}

func summaryMsg(files int, subject string, updated bool) string {
	action := stringTernary(updated, "removed", "obsolete")
	color := colorTernary(updated, greenText, yellowText)

	return color(fmt.Sprintf("%s%d snapshot %s %s.\n", arrowPoint, files, subject, action))
}

func stringTernary(assertion bool, trueBranch string, falseBranch string) string {
	if !assertion {
		return falseBranch
	}

	return trueBranch
}

func colorTernary(
	assertion bool,
	colorFuncTrue func(string) string,
	colorFuncFalse func(string) string,
) func(string) string {
	if !assertion {
		return colorFuncFalse
	}

	return colorFuncTrue
}

/*
	Builds a Set with all snapshot ids registered inside a snap file
	Form: testname - number id

	tests have the form
		map[filepath]: map[testname]: <number of snapshots>

	e.g
		./path/__snapshots__/add_test.snap map[TestAdd] 3

		will result to

		TestAdd - 1
		TestAdd - 2
		TestAdd - 3

	as it means there are 3 snapshots created inside TestAdd
*/
func occurrences(tests map[string]int) set {
	result := make(set)
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
