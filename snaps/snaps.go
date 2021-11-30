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
