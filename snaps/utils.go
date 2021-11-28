package snaps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	testsOccur = map[string]int{}
	config     = Config{
		snapsDirName: "__snapshots__",
		snapsExtName: "snap",
	}
	snapshotNotFound = errors.New("Snapshot not found")
)

func corruptedSnapshot(snap string) error {
	return fmt.Errorf("Snapshot %s is corrupted", snap)
}

func registerTest(tName string) int {
	if c, exists := testsOccur[tName]; exists {
		testsOccur[tName] = c + 1
		return c + 1
	}

	testsOccur[tName] = 1
	return 1
}

func (c *Config) snapshotFile(tName string) (*os.File, error) { // TODO: maybe mark it for read only
	cPath := callerPath()
	p, fName := c.snapshotDir(tName, cPath), c.snapshotFileName(tName, cPath)
	if err := os.MkdirAll(p, 0770); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(p, fName), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0770)
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

func callerPath() string {
	pc := make([]uintptr, 5)
	// with 0 identifying the frame for Callers itself and 1 identifying the caller of Callers
	// and we skip the "getTestFileInfo" call
	// FIXME: comment
	n := runtime.Callers(5, pc)

	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.File
}
