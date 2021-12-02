package snaps

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/kr/pretty"
)

var (
	// We track occurrence as in the same test we can run multiple snapshots
	testsOccur      = map[string]int{}
	errSnapNotFound = errors.New("snapshot not found")
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

// Returns the id of the test in the snapshot
// Form [<test-name> - <occurrence>]
func getTestID(tName string) string {
	occurrence := testsOccur[tName]
	return fmt.Sprintf("[%s - %d]", tName, occurrence)
}

// Returns the path where the "user" tests are running
func baseCaller() string {
	pc := make([]uintptr, 50) // NOTE: this might not be enough
	// with 0 identifying the frame for Callers itself and 1 identifying the caller of Callers
	n := runtime.Callers(0, pc)

	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()
	prevFile := frame.File

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
