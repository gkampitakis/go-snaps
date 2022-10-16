package snaps

import (
	"errors"
	"github.com/gkampitakis/go-snaps/snaps/internal/formatter"

	"github.com/gkampitakis/go-snaps/snaps/internal/colors"
)

// MatchSnapshot verifies the values match the most recent snap file
//
// You can pass multiple values
//
//	MatchSnapshot(t, 10, "hello world")
//
// or call MatchSnapshot multiples times inside a test
//
//	MatchSnapshot(t, 10)
//	MatchSnapshot(t, "hello world")
//
// The difference is the latter will create multiple entries.
func MatchSnapshot(t testingT, values ...interface{}) {
	t.Helper()

	if len(values) == 0 {
		t.Log(colors.Sprint(colors.Yellow, "[warning] MatchSnapshot call without params\n"))
		return
	}

	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)
	snapshot := takeSnapshot(values)
	prevSnapshot, err := getPrevSnapshot(testID, snapPath)

	if errors.Is(err, errSnapNotFound) {
		if isCI {
			handleError(t, err)
			return
		}

		err := addNewSnapshot(testID, snapshot, dir, snapPath)
		if err != nil {
			handleError(t, err)
			return
		}

		t.Log(addedMsg)
		testEvents.register(added)
		return
	}
	if err != nil {
		handleError(t, err)
		return
	}

	diff := prettyDiff(unescapeEndChars(prevSnapshot), unescapeEndChars(snapshot))
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate {
		handleError(t, diff)
		return
	}

	if err = updateSnapshot(testID, snapshot, snapPath); err != nil {
		handleError(t, err)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}

func takeSnapshot(objects []interface{}) string {
	var snapshot string

	for i := 0; i < len(objects); i++ {
		snapshot += formatter.Sprint(objects[i]) + "\n"
	}

	return escapeEndChars(snapshot)
}
