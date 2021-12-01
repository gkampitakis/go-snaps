package snaps

import (
	"testing"
)

func MatchSnapshot(t *testing.T, o ...interface{}) {
	defaultConfig.matchSnapshot(t, &o)
}

func (c *Config) MatchSnapshot(t *testing.T, o ...interface{}) {
	c.matchSnapshot(t, &o)
}

// NOTE: need to find better way for identifying caller
// NOTE: how can we identify CI
// NOTE: clear functions and folder names, what should go where
// NOTE: we need to test with table tests if snapshots work
// NOTE: can break tests if we pass something weird in re := regexp.MustCompile("(?:\\" + testID + "[\\s\\S])(.*[\\s\\S]*?)(?:---)")
// NOTE: Can we have race conditions if we don't lock file operations
// NOTE: can loading the whole file cause issues in updating snapshot
