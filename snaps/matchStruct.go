package snaps

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/match"
	"github.com/kr/pretty"
)

func (c *config) MatchStruct(t testingT, input interface{}, matchers ...match.StructMatcher) {
	t.Helper()

	matchStruct(c, t, input, matchers...)
}

func MatchStruct(t testingT, input interface{}, matchers ...match.StructMatcher) {
	t.Helper()

	matchStruct(&defaultConfig, t, input, matchers...)
}

func matchStruct(c *config, t testingT, input interface{}, matchers ...match.StructMatcher) {
	t.Helper()

	// TODO: add validation that input is indeed a struct

	dir, snapPath := snapDirAndName(c)
	testID := testsRegistry.getTestID(t.Name(), snapPath)

	s, matchersErrors := applyStructMatchers(input, matchers...)
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

	snapshot := pretty.Sprintf("%# v\n", s)
	prevSnapshot, err := getPrevSnapshot(testID, snapPath)
	if errors.Is(err, errSnapNotFound) {
		if isCI {
			handleError(t, err)
			return
		}

		err := addNewSnapshot(testID, snapshot, dir, snapPath)
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

	diff := prettyDiff(prevSnapshot, snapshot)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdateSingle(t.Name()) {
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

func applyStructMatchers(input interface{}, matchers ...match.StructMatcher) (interface{}, []match.MatcherError) {
	errors := []match.MatcherError{}

	for _, m := range matchers {
		_struct, errs := m.Struct(input)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}
		input = _struct
	}

	return input, errors
}
