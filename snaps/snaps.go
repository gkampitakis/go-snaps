package snaps

import (
	"testing"
)

func MatchSnapshot(t *testing.T, o ...interface{}) {
	t.Helper()

	defaultConfig.matchSnapshot(t, &o)
}

func (c *Config) MatchSnapshot(t *testing.T, o ...interface{}) {
	t.Helper()

	c.matchSnapshot(t, &o)
}
