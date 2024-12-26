package snaps

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/match"
)

func (c *Config) MatchStandaloneJSON(t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	if c.extension == "" {
		c.extension = ".json"
	}

	matchStandaloneJSON(c, t, input, matchers...)
}

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

	j, matchersErrors := applyJSONMatchers(j, matchers...)
	if len(matchersErrors) > 0 {
		s := strings.Builder{}

		for _, err := range matchersErrors {
			colors.Fprint(
				&s,
				colors.Red,
				fmt.Sprintf(
					"\n%smatch.%s(\"%s\") - %s",
					errorSymbol,
					err.Matcher,
					err.Path,
					err.Reason,
				),
			)
		}

		handleError(t, s.String())
		return
	}

	snapshot := takeJSONSnapshot(j)
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
		1,
	)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(c.update) {
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
