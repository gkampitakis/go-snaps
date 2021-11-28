package snaps

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/kr/pretty"
)

func MatchSnapshot(t *testing.T, o ...interface{}) {
	if len(o) == 0 {
		return
	}
	registerTest(t.Name())
	snap := takeSnapshot(&o)
	prevSnap, err := getPrevSnapshot(t.Name())
	if err != nil {
		t.Error(err)
	}

	if desc := pretty.Diff(snap, prevSnap); len(desc) > 0 {
		t.Error(desc)
		// FIXME:  how to change stack trace here
		// TODO: break it to a function
	}
}

func takeSnapshot(objects *[]interface{}) string {
	var snapshot string

	for i := 0; i < len(*objects); i++ {
		snapshot += pretty.Sprint((*objects)[0]) + "\n"
	}

	return snapshot
}

func getPrevSnapshot(name string) (string, error) {
	var snapshot string
	found, end := false, false
	occurrence := testsOccur[name]
	testID := fmt.Sprintf("[%s - %d]", name, occurrence)

	f, err := config.snapshotFile(name)
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

func saveSnapshot(f *os.File, snapshotName string, o []interface{}) error {
	// FIXME: refactor, use the take snapshot and then attach the identifier
	snapshot := fmt.Sprintf("[%s]\n%s\n---\n", snapshotName, pretty.Sprint(o[0]))

	_, err := f.WriteString(snapshot)
	return err
}
