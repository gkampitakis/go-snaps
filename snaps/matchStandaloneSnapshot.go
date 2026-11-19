package snaps

import (
	"github.com/kr/pretty"
)

/*
MatchStandaloneSnapshot verifies the input matches the most recent snap file

	MatchStandaloneSnapshot(t, "Hello World")

MatchStandaloneSnapshot creates one snapshot file per call.

You can call MatchStandaloneSnapshot multiple times inside a test.
It will create multiple snapshot files at `__snapshots__` folder by default.
*/
func (c *Config) MatchStandaloneSnapshot(t testingT, input any) {
	t.Helper()

	matchStandaloneSnapshot(c, t, input)
}

/*
MatchStandaloneSnapshot verifies the input matches the most recent snap file

	MatchStandaloneSnapshot(t, "Hello World")

MatchStandaloneSnapshot creates one snapshot file per call.

You can call MatchStandaloneSnapshot multiple times inside a test.
It will create multiple snapshot files at `__snapshots__` folder by default.
*/
func MatchStandaloneSnapshot(t testingT, input any) {
	t.Helper()

	matchStandaloneSnapshot(&defaultConfig, t, input)
}

func (c *Config) takeStandaloneSnapshot(input any) string {
	if c.serializer != nil {
		return c.serializer(input)
	}

	return pretty.Sprint(input)
}

func matchStandaloneSnapshot(c *Config, t testingT, input any) {
	t.Helper()

	genericPathSnap, genericSnapPathRel := snapshotPath(c, t.Name(), true)
	snapPath, snapPathRel := standaloneTestsRegistry.getTestID(genericPathSnap, genericSnapPathRel)
	t.Cleanup(func() {
		standaloneTestsRegistry.reset(genericPathSnap)
	})

	snapshot := c.takeStandaloneSnapshot(input)
	compareAndStore(c, t, snapshot, standaloneStore{snapPath, snapPathRel}, false)
}
