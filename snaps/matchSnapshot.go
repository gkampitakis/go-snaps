package snaps

import (
	"errors"

	"github.com/gkampitakis/go-snaps/snaps/internal/colors"
)

// MatchSnapshot verifies the values match the most recent snap file
//
// you can call MatchSnapshot multiples times inside a test
//
//	MatchSnapshot(t, 10)
//	MatchSnapshot(t, "hello world")
func MatchSnapshot(t testingT, value interface{}) {
	t.Helper()

	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)
	snapshot := takeSnapshot(value)
	prevSnapshot, err := getPrevSnapshot(testID, snapPath)

	if errors.Is(err, errSnapNotFound) {
		if isCI {
			t.Error(err)
			return
		}

		err := addNewSnapshot(testID, snapshot, dir, snapPath)
		if err != nil {
			t.Error(err)
			return
		}

		t.Log(colors.Sprint(colors.Green, arrowPoint+"New snapshot written.\n"))
		return
	}
	if err != nil {
		t.Error(err)
		return
	}

	diff := prettyDiff(unescapeEndChars(prevSnapshot), unescapeEndChars(snapshot))
	if diff == "" {
		return
	}

	if !shouldUpdate {
		t.Error(diff)
		return
	}

	if err = updateSnapshot(testID, snapshot, snapPath); err != nil {
		t.Error(err)
		return
	}

	t.Log(colors.Sprint(colors.Green, arrowPoint+"Snapshot updated.\n"))
}
