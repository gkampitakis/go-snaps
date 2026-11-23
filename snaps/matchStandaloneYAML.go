package snaps

import (
	"github.com/gkampitakis/go-snaps/match"
)

/*
MatchStandaloneYAML verifies the input matches the most recent snap file.
Input can be a valid yaml string or []byte or whatever value can be passed
successfully on `yaml.Marshal`.

	snaps.MatchStandaloneYAML(t, "user: \"mock-user\"\nage: 10\nemail: mock@email.com")
	snaps.MatchStandaloneYAML(t, []byte("user: \"mock-user\"\nage: 10\nemail: mock@email.com"))
	snaps.MatchStandaloneYAML(t, User{10, "mock-email"})

MatchStandaloneYAML also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	snaps.MatchStandaloneYAML(t, User{Created: time.Now(), Email: "mock-email"}, match.Any("$.created"))

MatchStandaloneYAML creates one snapshot file per call.

You can call MatchStandaloneYAML multiple times inside a test.
It will create multiple snapshot files at `__snapshots__` folder by default.
*/
func (c *Config) MatchStandaloneYAML(t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	if c.extension == "" {
		c.extension = ".yaml"
	}

	matchStandaloneYAML(c, t, input, matchers...)
}

/*
MatchStandaloneYAML verifies the input matches the most recent snap file.
Input can be a valid yaml string or []byte or whatever value can be passed
successfully on `yaml.Marshal`.

	snaps.MatchStandaloneYAML(t, "user: \"mock-user\"\nage: 10\nemail: mock@email.com")
	snaps.MatchStandaloneYAML(t, []byte("user: \"mock-user\"\nage: 10\nemail: mock@email.com"))
	snaps.MatchStandaloneYAML(t, User{10, "mock-email"})

MatchStandaloneYAML also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	snaps.MatchStandaloneYAML(t, User{Created: time.Now(), Email: "mock-email"}, match.Any("$.created"))

MatchStandaloneYAML creates one snapshot file per call.

You can call MatchStandaloneYAML multiple times inside a test.
It will create multiple snapshot files at `__snapshots__` folder by default.
*/
func MatchStandaloneYAML(t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	c := defaultConfig
	if c.extension == "" {
		c.extension = ".yaml"
	}

	matchStandaloneYAML(&c, t, input, matchers...)
}

func matchStandaloneYAML(c *Config, t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	genericPathSnap, genericSnapPathRel := snapshotPath(c, t.Name(), true)
	snapPath, snapPathRel := standaloneTestsRegistry.getTestID(genericPathSnap, genericSnapPathRel)
	t.Cleanup(func() {
		standaloneTestsRegistry.reset(genericPathSnap)
	})

	y, err := validateYAML(input)
	if err != nil {
		handleError(t, err)
		return
	}

	y, matchersErrors := applyMatchers(y, matchers, match.YAMLMatcher.YAML)
	if len(matchersErrors) > 0 {
		handleError(t, formatMatcherErrors(matchersErrors))
		return
	}

	snapshot := takeYAMLSnapshot(y)
	compareAndStore(c, t, snapshot, standaloneStore{snapPath, snapPathRel}, false)
}
