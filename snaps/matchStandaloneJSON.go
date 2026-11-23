package snaps

import (
	"github.com/gkampitakis/go-snaps/match"
)

/*
MatchStandaloneJSON verifies the input matches the most recent snap file.
Input can be a valid json string or []byte or whatever value can be passed
successfully on `json.Marshal`.

	snaps.MatchStandaloneJSON(t, `{"user":"mock-user","age":10,"email":"mock@email.com"}`)
	snaps.MatchStandaloneJSON(t, []byte(`{"user":"mock-user","age":10,"email":"mock@email.com"}`))
	snaps.MatchStandaloneJSON(t, User{10, "mock-email"})

MatchStandaloneJSON also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	snaps.MatchStandaloneJSON(t, User{Created: time.Now(), Email: "mock-email"}, match.Any("created"))

MatchStandaloneJSON creates one snapshot file per call.

You can call MatchStandaloneJSON multiple times inside a test.
It will create multiple snapshot files at `__snapshots__` folder by default.
*/
func (c *Config) MatchStandaloneJSON(t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	if c.extension == "" {
		c.extension = ".json"
	}

	matchStandaloneJSON(c, t, input, matchers...)
}

/*
MatchStandaloneJSON verifies the input matches the most recent snap file.
Input can be a valid json string or []byte or whatever value can be passed
successfully on `json.Marshal`.

	snaps.MatchStandaloneJSON(t, `{"user":"mock-user","age":10,"email":"mock@email.com"}`)
	snaps.MatchStandaloneJSON(t, []byte(`{"user":"mock-user","age":10,"email":"mock@email.com"}`))
	snaps.MatchStandaloneJSON(t, User{10, "mock-email"})

MatchStandaloneJSON also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	snaps.MatchStandaloneJSON(t, User{Created: time.Now(), Email: "mock-email"}, match.Any("created"))

MatchStandaloneJSON creates one snapshot file per call.

You can call MatchStandaloneJSON multiple times inside a test.
It will create multiple snapshot files at `__snapshots__` folder by default.
*/
func MatchStandaloneJSON(t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	c := defaultConfig
	if c.extension == "" {
		c.extension = ".json"
	}

	matchStandaloneJSON(&c, t, input, matchers...)
}

func matchStandaloneJSON(c *Config, t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	genericPathSnap, genericSnapPathRel := snapshotPath(c, t.Name(), true)
	snapPath, snapPathRel := standaloneTestsRegistry.getTestID(genericPathSnap, genericSnapPathRel)
	t.Cleanup(func() {
		standaloneTestsRegistry.reset(genericPathSnap)
	})

	j, err := validateJSON(input)
	if err != nil {
		handleError(t, err)
		return
	}

	j, matchersErrors := applyMatchers(j, matchers, match.JSONMatcher.JSON)
	if len(matchersErrors) > 0 {
		handleError(t, formatMatcherErrors(matchersErrors))
		return
	}

	snapshot := c.takeJSONSnapshot(j)
	compareAndStore(c, t, snapshot, standaloneStore{snapPath, snapPathRel}, false)
}
