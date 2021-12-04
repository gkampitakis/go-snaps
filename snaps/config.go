package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type Config struct {
	snapsDir     string
	snapsExt     string
	shouldUpdate bool
}

var defaultConfig = Config{
	snapsDir:     "__snapshots__",
	snapsExt:     "snap",
	shouldUpdate: getEnvBool("UPDATE_SNAPS", false),
}

// Get a new snaps instance configured
func New(configurers ...func(*Config)) *Config {
	config := defaultConfig

	for _, c := range configurers {
		c(&config)
	}

	return &config
}

// Set snaps directory. Default "__snapshots__"
func SnapsDirectory(dir string) func(*Config) {
	return func(c *Config) {
		c.snapsDir = dir
	}
}

// Set snap file extension. Default "snap"
func SnapsExtension(extension string) func(*Config) {
	return func(c *Config) {
		c.snapsExt = extension
	}
}

// Using pointer here so we can remove the interface{} when printing
func (c *Config) matchSnapshot(t *testing.T, o *[]interface{}) {
	t.Helper()

	if len(*o) == 0 {
		return
	}

	path, fPath := c.snapPathAndFile()
	testID := getTestID(t.Name())
	snap := takeSnapshot(o)
	prevSnap, err := c.getPrevSnapshot(testID, fPath)

	if errors.Is(err, errSnapNotFound) {
		err := c.addNewSnapshot(testID, snap, path, fPath)
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
			err := c.updateSnapshot(testID, snap, fPath)
			if err != nil {
				t.Error(err)
			}

			fmt.Print(greenText("1 snapshot was updated\n"))
			return
		}

		t.Error(diff)
	}
}

func (c *Config) getPrevSnapshot(testID, fPath string) (string, error) {
	f, err := c.snapshotFileToString(fPath)
	if err != nil {
		return "", err
	}

	match := testIDRegex(testID).FindStringSubmatch(f)

	if len(match) < 2 {
		return "", errSnapNotFound
	}

	// The second capture group has the snapshot data
	return match[1], err
}

func (c *Config) snapshotFileToString(fPath string) (string, error) {
	_, err := os.Stat(fPath)
	if err != nil {
		return "", errSnapNotFound
	}

	f, err := os.ReadFile(fPath)
	if err != nil {
		return "", err
	}

	return string(f), err
}

func (c *Config) stringToSnapshotFile(snap, fPath string) error {
	return os.WriteFile(fPath, []byte(snap), os.ModePerm)
}

func (c *Config) addNewSnapshot(testID, snap, path, fPath string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(fPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("\n%s\n%s---\n", testID, snap))
	if err != nil {
		return err
	}

	return nil
}

/*
	Returns the path (p) where the tests run + /snapsDirName
	and the second string (f) is the the path + /snapsDirName + /<test-name>.snapsExtName
*/
func (c *Config) snapPathAndFile() (p, f string) {
	callerPath := baseCaller()
	base := filepath.Base(callerPath)

	p = filepath.Join(filepath.Dir(callerPath), c.snapsDir)
	f = filepath.Join(p, strings.TrimSuffix(base, filepath.Ext(base))+"."+c.snapsExt)

	return
}

func (c *Config) updateSnapshot(testID, snap, fPath string) error {
	// When t.Parallel a test can override another snapshot as we dump
	// all snapshots
	_m.Lock()
	defer _m.Unlock()
	f, err := c.snapshotFileToString(fPath)
	if err != nil {
		return err
	}

	newSnap := testIDRegex(testID).
		ReplaceAllString(f, fmt.Sprintf("%s\n%s---", testID, snap))

	err = c.stringToSnapshotFile(newSnap, fPath)
	if err != nil {
		return err
	}

	return nil
}
