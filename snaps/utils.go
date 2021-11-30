package snaps

import (
	"errors"
	"fmt"
	"runtime"
)

var (
	testsOccur       = map[string]int{}
	snapshotNotFound = errors.New("Snapshot not found")
)

const (
	resetCode   = "\u001b[0m"
	redBGCode   = "\u001b[41m\u001b[37;1m"
	greenBGCode = "\u001b[42m\u001b[37;1m"
	dimCode     = "\u001b[2m"
	greenCode   = "\u001b[32m"
)

func redBG(txt string) string {
	return fmt.Sprintf("%s%s%s", redBGCode, txt, resetCode)
}

func greenBG(txt string) string {
	return fmt.Sprintf("%s%s%s", greenBGCode, txt, resetCode)
}

func dimText(txt string) string {
	return fmt.Sprintf("%s%s%s", dimCode, txt, resetCode)
}

func greenText(txt string) string {
	return fmt.Sprintf("%s%s%s", greenCode, txt, resetCode)
}

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

func callerPath() string {
	// NOTE: we need to make this more smart
	pc := make([]uintptr, 6)
	// with 0 identifying the frame for Callers itself and 1 identifying the caller of Callers
	// and we skip the "getTestFileInfo" call
	// FIXME: comment
	n := runtime.Callers(6, pc)

	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	return frame.File
}
