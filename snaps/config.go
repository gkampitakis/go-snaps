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
	registerTest(t.Name())
	snap := takeSnapshot(o)
	prevSnap, err := c.getPrevSnapshot(t.Name(), fPath)

	if errors.Is(err, errSnapNotFound) {
		err := c.addNewSnapshot(t.Name(), snap, path, fPath)
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
			err := c.updateSnapshot(t.Name(), snap, fPath)
			if err != nil {
				t.Error(err)
			}

			fmt.Print(greenText("1 snapshot was updated\n"))
			return
		}

		t.Error(diff)
	}
}

func (c *Config) getPrevSnapshot(tName, fPath string) (string, error) {
	f, err := c.snapshotFileToString(tName, fPath)
	if err != nil {
		return "", err
	}

	testID := getTestID(tName)

	// e.g (?:\[TestAdd\/Hello_World\/my-test - 1\][\s\S])(.*[\s\S]*?)(?:---)
	re := regexp.MustCompile("(?:\\" + testID + "[\\s\\S])(.*[\\s\\S]*?)(?:---)")
	match := re.FindStringSubmatch(f)

	if len(match) < 2 {
		return "", errSnapNotFound
	}

	// The second capture group has the snapshot data
	return match[1], err
}

func (c *Config) snapshotFileToString(tName, fPath string) (string, error) {
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

func (c *Config) stringToSnapshotFile(tName, snap, fPath string) error {
	return os.WriteFile(fPath, []byte(snap), os.ModePerm)
}

func (c *Config) addNewSnapshot(tName, snap, path, fPath string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(fPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	testID := getTestID(tName)
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

func (c *Config) updateSnapshot(tName, snap, fPath string) error {
	f, err := c.snapshotFileToString(tName, fPath)
	if err != nil {
		return err
	}

	testID := getTestID(tName)

	re := regexp.MustCompile("(?:\\" + testID + "[\\s\\S])(.*[\\s\\S]*?)(?:---)")
	newSnap := re.ReplaceAllString(f, fmt.Sprintf("%s\n%s---", testID, snap))

	err = c.stringToSnapshotFile(tName, newSnap, fPath)
	if err != nil {
		return err
	}

	return nil
}
