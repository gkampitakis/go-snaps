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
