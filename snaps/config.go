package snaps

import (
	"os"
	"path/filepath"
	"strings"
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
		shouldUpdate: false,
	}
}

var defaultConfig = Config{
	snapsDirName: "__snapshots__",
	snapsExtName: "snap",
	shouldUpdate: getEnvBool("UPDATE_SNAPS", false),
}

// FIXME: rewrite better and declutter repeated code
func (c *Config) snapshotFileString(tName string) (string, error) { // to string
	cPath := callerPath() // NOTE: this can be cached in another level if all functions call from the same depth
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	if err := os.MkdirAll(p, 0770); err != nil {
		return "", err
	}
	f, err := os.ReadFile(filepath.Join(p, fName))
	if err != nil {
		return "", err
	}

	return string(f), err
}

func (c *Config) stringToSnapshot(tName, snap string) error {
	cPath := callerPath() // NOTE: this can be cached in another level if all functions call from the same depth
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	if err := os.MkdirAll(p, 0770); err != nil {
		return err
	}
	err := os.WriteFile(filepath.Join(p, fName), []byte(snap), 0770)
	if err != nil {
		return err
	}

	return err
}

func (c *Config) snapshotFile(tName string, rw int) (*os.File, error) {
	cPath := callerPath() //FIXME: we need to groom this and make it generic for stack error trace
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	if err := os.MkdirAll(p, 0770); err != nil {
		return nil, err
	}

	// FIXME: we need to investigate this one
	file, err := os.OpenFile(filepath.Join(p, fName), os.O_APPEND|os.O_CREATE|rw, 0770)
	if err != nil {
		return nil, err
	}

	return file, err
}

func (c *Config) snapshotDir(tName, cPath string) string {
	return filepath.Join(filepath.Dir(cPath), c.snapsDirName)
}

func (c *Config) snapshotFileName(tName, cPath string) string {
	base := filepath.Base(cPath)
	return strings.TrimSuffix(base, filepath.Ext(base)) + "." + c.snapsExtName
}
