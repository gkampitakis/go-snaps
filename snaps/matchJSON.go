package snaps

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

/*
MatchJSON verifies the input matches the most recent snap file.
Input can be a valid json string or []byte or a struct that can marshalled to a json

MatchJSON also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	MatchJSON(`{"user":{"bday":"10/10/2010","name":"gkampitakis"}}`, match.Any("user.bday"))
*/
func MatchJSON(t testingT, input interface{}, matchers ...match.JSONMatcher) {
	t.Helper()
	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)

	j, err := validateJSON(input)
	if err != nil {
		handleError(t, err)
		return
	}

	j, matchersErrors := applyMatchers(j, matchers...)
	if len(matchersErrors) > 0 {
		s := "\nMatchers failed\n"
		for _, err := range matchersErrors {
			s += fmt.Sprintf("- match.%s(\"%s\") x %s\n", err.Matcher, err.Path, err.Reason)
		}

		handleError(t, s)
		return
	}

	snapshot := takeJSONSnapshot(j)
	if err != nil {
		handleError(t, err)
		return
	}
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

	diff := prettyDiff(unescapeEndChars(prevSnapshot), unescapeEndChars(snapshot))
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate {
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

func validateJSON(input interface{}) ([]byte, error) {
	switch j := input.(type) {
	case string:
		if !gjson.Valid(j) {
			return nil, errInvalidJSON
		}

		return []byte(j), nil
	case []byte:
		if !gjson.ValidBytes(j) {
			return nil, errInvalidJSON
		}

		return j, nil
	default:
		return json.Marshal(input)
	}
}

func takeJSONSnapshot(b []byte) string {
	return string(pretty.PrettyOptions(b, jsonOptions))
}

func applyMatchers(b []byte, matchers ...match.JSONMatcher) ([]byte, []match.MatcherError) {
	errors := []match.MatcherError{}

	for _, m := range matchers {
		json, errs := m.JSON(b)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}
		b = json
	}

	return b, errors
}
