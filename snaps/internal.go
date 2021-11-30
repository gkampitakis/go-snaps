package snaps

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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
		err := c.saveSnapshot(snap, t.Name())
		if err != nil {
			t.Error(err)
		}

		fmt.Print(greenText("1 snapshot was written\n"))
		return
	}
	if err != nil {
		t.Error(err)
	}

	prettyPrintDiff(t, prevSnap, snap)
}

func takeSnapshot(objects *[]interface{}) string {
	var snapshot string

	for i := 0; i < len(*objects); i++ {
		snapshot += pretty.Sprint((*objects)[i]) + "\n"
	}

	return snapshot
}

func (c *Config) getPrevSnapshot(name string) (string, error) {
	var snapshot string
	found, end := false, false
	occurrence := testsOccur[name]
	testID := fmt.Sprintf("[%s - %d]", name, occurrence)

	f, err := c.snapshotFile(name, os.O_RDONLY)
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

func (c *Config) saveSnapshot(snapshot, name string) error {
	occurrence := testsOccur[name]
	f, err := c.snapshotFile(name, os.O_RDWR)
	if err != nil {
		return err
	}
	defer f.Close()

	s := fmt.Sprintf("\n[%s - %d]\n%s---\n", name, occurrence, snapshot)

	_, err = f.WriteString(s)
	return err
}
