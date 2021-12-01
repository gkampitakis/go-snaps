package snaps

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/kr/pretty"
)

func (c *Config) matchSnapshot(t *testing.T, o *[]interface{}) {
	if len(*o) == 0 {
		return
	}
	registerTest(t.Name())
	snap := takeSnapshot(o)
	prevSnap, err := c.getPrevSnapshot(t.Name())

	if errors.Is(err, snapshotNotFound) {
		err := c.saveSnapshot(t.Name(), snap)
		if err != nil {
			t.Error(err)
		}

		fmt.Print(greenText("1 snapshot was written\n"))
		return
	}
	if err != nil {
		t.Error(err)
	}

	diff := prettyPrintDiff(t, prevSnap, snap)
	if diff != "" {

		if c.shouldUpdate {
			fmt.Print(greenText("1 snapshot was updated\n"))
			// NOTE: can this be problematic if we don't lock file ?? This will be investigated last
			// NOTE: can this be an issue in cases of big snapshot files ?? can do something smarter ?
			err := c.updateSnapshot(t.Name(), snap)
			if err != nil {
				t.Error(err)
			}

			return
		}

		fmt.Print(diff)
		t.Error("diffs")

		// FIXME: error message
		// FIXME: how to change stack trace here
	}
}

func (c *Config) updateSnapshot(tName, snap string) error {
	testID := getTestID(tName)
	f, err := c.snapshotFileString(tName)
	if err != nil {
		return err
	}

	// (?:\[TestAdd\/Hello_World\/more_tests - 1\][\s\S]).*[\s\S]*?(?:---)
	// (?<=\[TestAdd\/Hello_World\/more_tests - 1\][\s\S]).*[\s\S]*?(?=---)

	// NOTE: can this broke tests if we pass something weird here?
	re := regexp.MustCompile("(?:\\" + testID + "[\\s\\S])(.*[\\s\\S]*?)(?:---)") // NOTE: this removes the test id, if we can remove it then we might
	// can refactor the get prev snapsot?
	// FIXME: right now prevSnapshot is loading the file line by line as the update loads the whole file, we need to stick to one way. Try the most optimal
	// The preferred would be open it once and load it gradually ? needs more thinking and investigation
	newSnap := re.ReplaceAllString(f, snapshot(snap, testID))

	c.stringToSnapshot(tName, newSnap)
	if err != nil {
		return err
	}

	return nil
}

func takeSnapshot(objects *[]interface{}) string {
	var snapshot string

	for i := 0; i < len(*objects); i++ {
		snapshot += pretty.Sprint((*objects)[i]) + "\n"
	}

	return snapshot
}

func (c *Config) getPrevSnapshot(tName string) (string, error) {
	var snapshot string
	found, end := false, false
	testID := getTestID(tName)

	f, err := c.snapshotFile(tName, os.O_RDONLY) // TODO: we need to add a lock here for write in update
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()

		if text == testID {
			found = true
			continue
		}

		if found && text == "---" {
			end = true
			break
		}

		if found {
			snapshot += text + "\n"
		}
	}

	if !found {
		return "", snapshotNotFound
	}

	if !end {
		return "", corruptedSnapshot(testID)
	}

	return snapshot, nil
}

func snapshot(snap, testID string) string {
	return fmt.Sprintf("\n%s\n%s---\n", testID, snap)
}

func getTestID(tName string) string {
	occurrence := testsOccur[tName]
	return fmt.Sprintf("[%s - %d]", tName, occurrence)
}

func (c *Config) saveSnapshot(tName, snap string) error {
	testID := getTestID(tName)
	f, err := c.snapshotFile(tName, os.O_RDWR)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(snapshot(snap, testID))

	return err
}
