package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type Config struct {
	snapsDirName string
	snapsExtName string
	shouldUpdate bool
}

func New() *Config {
	// FIXME:
	return &Config{
		snapsDirName: "snapshots",
		snapsExtName: "snap",
		shouldUpdate: getEnvBool("UPDATE_SNAPS", false),
	}
}

var defaultConfig = Config{
	snapsDirName: "__snapshots__",
	snapsExtName: "snap",
	shouldUpdate: getEnvBool("UPDATE_SNAPS", false),
}

func (c *Config) matchSnapshot(t *testing.T, o *[]interface{}) {
	if len(*o) == 0 {
		return
	}

	registerTest(t.Name())
	snap := takeSnapshot(o)
	prevSnap, err := c.getPrevSnapshot(t.Name())

	if errors.Is(err, snapshotNotFound) {
		err := c.addNewSnapshot(t.Name(), snap)
		if err != nil {
			t.Error(err)
		}

		fmt.Print(greenText("1 snapshot was written\n"))
		return
	}
	if err != nil {
		t.Error(err)
	}

	diff := prettyDiff(prevSnap, snap)
	if diff != "" {
		if c.shouldUpdate {
			err := c.updateSnapshot(t.Name(), snap)
			if err != nil {
				t.Error(err)
			}

			fmt.Print(greenText("1 snapshot was updated\n"))
			return
		}

		fmt.Print(diff)
		t.Error("diffs")

		// FIXME: error message
		// FIXME: how to change stack trace here
	}
}

func (c *Config) getPrevSnapshot(tName string) (string, error) {
	testID := getTestID(tName)
	f, err := c.snapshotFileToString(tName)
	if err != nil {
		return "", err
	}

	// e.g (?:\[TestAdd\/Hello_World\/my-test - 1\][\s\S])(.*[\s\S]*?)(?:---)
	re := regexp.MustCompile("(?:\\" + testID + "[\\s\\S])(.*[\\s\\S]*?)(?:---)")
	match := re.FindStringSubmatch(f)

	if len(match) < 2 {
		return "", snapshotNotFound
	}

	// The second capture group has the snapshot data
	return match[1], err
}

func (c *Config) snapshotFileToString(tName string) (string, error) {
	cPath := baseCaller()
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	fPath := filepath.Join(p, fName)

	_, err := os.Stat(fPath)
	if err != nil {
		return "", snapshotNotFound
	}

	f, err := os.ReadFile(fPath)
	if err != nil {
		return "", err
	}

	return string(f), err
}

func (c *Config) stringToSnapshotFile(tName, snap string) error {
	cPath := baseCaller()
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)

	err := os.WriteFile(filepath.Join(p, fName), []byte(snap), os.ModePerm)
	if err != nil {
		return err
	}

	return err
}

func (c *Config) addNewSnapshot(tName, snap string) error {
	testID := getTestID(tName)
	cPath := baseCaller()
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(p, fName), os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(snapshotEntry(snap, testID))
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) snapshotDir(tName, cPath string) string {
	return filepath.Join(filepath.Dir(cPath), c.snapsDirName)
}

func (c *Config) snapshotFileName(tName, cPath string) string {
	base := filepath.Base(cPath)
	return strings.TrimSuffix(base, filepath.Ext(base)) + "." + c.snapsExtName
}

func (c *Config) updateSnapshot(tName, snap string) error {
	// BUG: adds one new line above and below
	testID := getTestID(tName)
	f, err := c.snapshotFileToString(tName)
	if err != nil {
		return err
	}

	re := regexp.MustCompile("(?:\\" + testID + "[\\s\\S])(.*[\\s\\S]*?)(?:---)")
	newSnap := re.ReplaceAllString(f, snapshotEntry(snap, testID))

	c.stringToSnapshotFile(tName, newSnap)
	if err != nil {
		return err
	}

	return nil
}
