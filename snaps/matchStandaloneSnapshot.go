package snaps

import (
	"errors"

	"github.com/kr/pretty"
)

func (c *config) MatchStandaloneSnapshot(t testingT, value any) {
	t.Helper()

	matchStandaloneSnapshot(c, t, value)
}

func MatchStandaloneSnapshot(t testingT, value any) {
	t.Helper()

	matchStandaloneSnapshot(&defaultConfig, t, value)
}

func matchStandaloneSnapshot(c *config, t testingT, value any) {
	t.Helper()

	genericPathSnap, genericSnapPathRel := snapshotPath(c, t.Name(), true)
	snapPath, snapPathRel := standaloneTestsRegistry.getTestID(genericPathSnap, genericSnapPathRel)
	t.Cleanup(func() {
		standaloneTestsRegistry.reset(genericPathSnap)
	})

	snapshot := takeStandaloneSnapshot(value)
	prevSnapshot, err := getPrevStandaloneSnapshot(snapPath)
	if errors.Is(err, errSnapNotFound) {
		if isCI {
			handleError(t, err)
			return
		}

		err := upsertStandaloneSnapshot(snapshot, snapPath)
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

	diff := prettyDiff(
		prevSnapshot,
		snapshot,
		snapPathRel,
		0,
	)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(nil) {
		handleError(t, diff)
		return
	}

	if err = upsertStandaloneSnapshot(snapshot, snapPath); err != nil {
		handleError(t, err)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}

func takeStandaloneSnapshot(data any) string {
	return pretty.Sprint(data)
}
