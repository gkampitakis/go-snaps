package snaps

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	snapsDirName string
	snapsExtName string
}

func New() *Config {
	// FIXME:
	return &Config{
		snapsDirName: "snapshots",
		snapsExtName: "snap",
	}
}

var defaultConfig = Config{
	snapsDirName: "__snapshots__",
	snapsExtName: "snap",
}

func (c *Config) snapshotFile(tName string, rw int) (*os.File, error) {
	cPath := callerPath()
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	if err := os.MkdirAll(p, 0770); err != nil {
		return nil, err
	}

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
