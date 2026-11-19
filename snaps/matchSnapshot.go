package snaps

import (
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
	testID := testsRegistry.getTestID(snapPath, t.Name(), c.label)
	t.Cleanup(func() {
		testsRegistry.reset(snapPath, t.Name())
	})

	snapshot := c.takeSnapshot(values)
	compareAndStore(c, t, snapshot, snapFileStore{testID, snapPath, snapPathRel}, true)
}

func (c *Config) takeSnapshot(objects []any) string {
	snapshots := make([]string, len(objects))

	for i, object := range objects {
		if c.serializer != nil {
			snapshots[i] = c.serializer(object)
		} else {
			snapshots[i] = pretty.Sprint(object)
		}
	}

	return escapeEndChars(strings.Join(snapshots, "\n"))
}
