package snaps

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/match"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

var (
	jsonOptions = &pretty.Options{
		SortKeys: true,
		Indent:   " ",
	}
	errInvalidJSON = errors.New("invalid json")
)

/*
MatchJSON verifies the input matches the most recent snap file.
Input can be a valid json string or []byte or whatever value can be passed
successfully on `json.Marshal`.

	MatchJSON(t, `{"user":"mock-user","age":10,"email":"mock@email.com"}`)
	MatchJSON(t, []byte(`{"user":"mock-user","age":10,"email":"mock@email.com"}`))
	MatchJSON(t, User{10, "mock-email"})

MatchJSON also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	MatchJSON(t, User{created: time.Now(), email: "mock-email"}, match.Any("created"))
*/
func (c *Config) MatchJSON(t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	matchJSON(c, t, input, matchers...)
}

/*
MatchJSON verifies the input matches the most recent snap file.
Input can be a valid json string or []byte or whatever value can be passed
successfully on `json.Marshal`.

	MatchJSON(t, `{"user":"mock-user","age":10,"email":"mock@email.com"}`)
	MatchJSON(t, []byte(`{"user":"mock-user","age":10,"email":"mock@email.com"}`))
	MatchJSON(t, User{10, "mock-email"})

MatchJSON also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	MatchJSON(t, User{created: time.Now(), email: "mock-email"}, match.Any("created"))
*/
func MatchJSON(t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	matchJSON(&defaultConfig, t, input, matchers...)
}

func matchJSON(c *Config, t testingT, input any, matchers ...match.JSONMatcher) {
	t.Helper()

	snapPath, snapPathRel := snapshotPath(c, t.Name(), false)
	testID := testsRegistry.getTestID(snapPath, t.Name())
	t.Cleanup(func() {
		testsRegistry.reset(snapPath, t.Name())
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
	prevSnapshot, line, err := getPrevSnapshot(testID, snapPath)
	if errors.Is(err, errSnapNotFound) {
		if isCI {
			handleError(t, err)
			return
		}

		err := addNewSnapshot(testID, snapshot, snapPath)
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

	diff := prettyDiff(prevSnapshot, snapshot, snapPathRel, line)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(c.update) {
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

func validateJSON(input any) ([]byte, error) {
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
	return strings.TrimSuffix(string(pretty.PrettyOptions(b, jsonOptions)), "\n")
}

func applyJSONMatchers(b []byte, matchers ...match.JSONMatcher) ([]byte, []match.MatcherError) {
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
