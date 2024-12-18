package snaps

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/match"
	"github.com/goccy/go-yaml"
)

func (c *Config) MatchYAML(t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	matchYAML(c, t, input, matchers...)
}

func MatchYAML(t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	matchYAML(&defaultConfig, t, input, matchers...)
}

func matchYAML(c *Config, t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	snapPath, snapPathRel := snapshotPath(c, t.Name(), false)
	testID := testsRegistry.getTestID(snapPath, t.Name())
	t.Cleanup(func() {
		testsRegistry.reset(snapPath, t.Name())
	})

	y, err := validateYAML(input)
	if err != nil {
		handleError(t, err)
		return
	}

	y, matchersErrors := applyYAMLMatchers(y, matchers...)
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

	snapshot := string(y)
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

// FIXME: validate
func validateYAML(input any) ([]byte, error) {
	switch y := input.(type) {
	case string:
		return []byte(y), nil
	case []byte:
		return y, nil
	default:
		return yaml.Marshal(input)
	}
}

// NOTE: this can be improved and pass the ast file
// NOTE: technically here we want to arrive when the []byte is valid yaml
func applyYAMLMatchers(b []byte, matchers ...match.YAMLMatcher) ([]byte, []match.MatcherError) {
	errors := []match.MatcherError{}

	for _, m := range matchers {
		json, errs := m.YAML(b)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}
		b = json
	}

	return b, errors
}
