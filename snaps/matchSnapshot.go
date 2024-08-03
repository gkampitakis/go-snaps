package snaps

import (
	"errors"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/kr/pretty"
)

/*
MatchSnapshot verifies the values match the most recent snap file
You can pass multiple values

	MatchSnapshot(t, 10, "hello world")

or call MatchSnapshot multiples times inside a test

	MatchSnapshot(t, 10)
	MatchSnapshot(t, "hello world")

The difference is the latter will create multiple entries.
*/
func (c *Config) MatchSnapshot(t testingT, values ...any) {
	t.Helper()

	matchSnapshot(c, t, values...)
}

/*
MatchSnapshot verifies the values match the most recent snap file
You can pass multiple values

	MatchSnapshot(t, 10, "hello world")

or call MatchSnapshot multiples times inside a test

	MatchSnapshot(t, 10)
	MatchSnapshot(t, "hello world")

The difference is the latter will create multiple entries.
*/
func MatchSnapshot(t testingT, values ...any) {
	t.Helper()

	matchSnapshot(&defaultConfig, t, values...)
}

func matchSnapshot(c *Config, t testingT, values ...any) {
	t.Helper()

	if len(values) == 0 {
		t.Log(colors.Sprint(colors.Yellow, "[warning] MatchSnapshot call without params\n"))
		return
	}

	snapPath, snapPathRel := snapshotPath(c, t.Name(), false)
	testID := testsRegistry.getTestID(snapPath, t.Name())
	t.Cleanup(func() {
		testsRegistry.reset(snapPath, t.Name())
	})

	snapshot := takeSnapshot(values)
	prevSnapshot, line, err := getPrevSnapshot(testID, snapPath)
	if errors.Is(err, errSnapNotFound) {
		if isCI {
			handleError(t, err)
			return
		}

		err := addNewSnapshot(testID, snapshot, snapPath)
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
		unescapeEndChars(prevSnapshot),
		unescapeEndChars(snapshot),
		snapPathRel,
		line,
	)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(c.update) {
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

func takeSnapshot(objects []any) string {
	snapshots := make([]string, len(objects))

	for i, object := range objects {
		snapshots[i] = pretty.Sprint(object)
	}

	return escapeEndChars(strings.Join(snapshots, "\n"))
}
