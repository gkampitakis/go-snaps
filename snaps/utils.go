package snaps

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/kr/pretty"
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

func registerTest(tName string) int {
	if c, exists := testsOccur[tName]; exists {
		testsOccur[tName] = c + 1
		return c + 1
	}

	testsOccur[tName] = 1
	return 1
}

func takeSnapshot(objects *[]interface{}) string {
	var snapshot string

	for i := 0; i < len(*objects); i++ {
		snapshot += pretty.Sprint((*objects)[i]) + "\n"
	}

	return snapshot
}

func getTestID(tName string) string { // NOTE: this can be written better
	occurrence := testsOccur[tName]
	return fmt.Sprintf("[%s - %d]", tName, occurrence)
}

func snapshotEntry(snap, testID string) string {
	return fmt.Sprintf("\n%s\n%s---\n", testID, snap)
}

func baseCaller() string {
	pc := make([]uintptr, 50) // NOTE: this might not be enough
	// with 0 identifying the frame for Callers itself and 1 identifying the caller of Callers
	n := runtime.Callers(2, pc)

	frames := runtime.CallersFrames(pc[:n])
	var more = true
	var frame runtime.Frame
	prevFile := ""

	for more {
		prevFile = frame.File
		frame, more = frames.Next()
		tmp := strings.Split(frame.Function, ".")
		fName := tmp[len(tmp)-1]

		if fName == "tRunner" {
			break
		}
	}

	return prevFile
}

func callerStack() []string {
	pc := make([]uintptr, 50) // NOTE: this might not be enough
	// with 0 identifying the frame for Callers itself and 1 identifying the caller of Callers
	n := runtime.Callers(2, pc)

	frames := runtime.CallersFrames(pc[:n])
	var more = true
	var frame runtime.Frame
	callStack := []string{}

	for more {
		frame, more = frames.Next()
		tmp := strings.Split(frame.Function, ".")
		fName := tmp[len(tmp)-1]

		if fName == "tRunner" {
			break
		}

		callStack = append(callStack, fmt.Sprintf("%s:%d", fName, frame.Line))
	}

	return callStack
}
