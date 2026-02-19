package snaps

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/match"
	"github.com/goccy/go-yaml"
)

var yamlEncodeOptions = []yaml.EncodeOption{
	yaml.Indent(2),
	yaml.IndentSequence(true),
}

/*
MatchYAML verifies the input matches the most recent snap file.
Input can be a valid yaml string or []byte or whatever value can be passed
successfully on `yaml.Marshal`.

	snaps.MatchYAML(t, "user: \"mock-user\"\nage: 10\nemail: mock@email.com")
	snaps.MatchYAML(t, []byte("user: \"mock-user\"\nage: 10\nemail: mock@email.com"))
	snaps.MatchYAML(t, User{10, "mock-email"})

MatchYAML also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	snaps.MatchYAML(t, User{Created: time.Now(), Email: "mock-email"}, match.Any("$.created"))
*/
func (c *Config) MatchYAML(t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	matchYAML(c, t, input, matchers...)
}

/*
MatchYAML verifies the input matches the most recent snap file.
Input can be a valid yaml string or []byte or whatever value can be passed
successfully on `yaml.Marshal`.

	snaps.MatchYAML(t, "user: \"mock-user\"\nage: 10\nemail: mock@email.com")
	snaps.MatchYAML(t, []byte("user: \"mock-user\"\nage: 10\nemail: mock@email.com"))
	snaps.MatchYAML(t, User{10, "mock-email"})

MatchYAML also supports passing matchers as a third argument. Those matchers can act either as
validators or placeholders for data that might change on each invocation e.g. dates.

	snaps.MatchYAML(t, User{Created: time.Now(), Email: "mock-email"}, match.Any("$.created"))
*/
func MatchYAML(t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	matchYAML(&defaultConfig, t, input, matchers...)
}

func matchYAML(c *Config, t testingT, input any, matchers ...match.YAMLMatcher) {
	t.Helper()

	snapPath, snapPathRel := snapshotPath(c, t.Name(), false)
	testID := testsRegistry.getTestID(snapPath, t.Name(), c.label)
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

	snapshot := takeYAMLSnapshot(y)
	prevSnapshot, line, err := getPrevSnapshot(testID, snapPath)
	if errors.Is(err, errSnapNotFound) {
		if !shouldCreate(c.update) {
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

	diff := prettyDiff(
		unescapeEndChars(prevSnapshot),
		unescapeEndChars(snapshot),
		snapPathRel,
		line,
	)
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

func validateYAML(input any) ([]byte, error) {
	var out any

	switch y := input.(type) {
	case string:
		err := yaml.Unmarshal([]byte(y), &out)
		if err != nil {
			return nil, fmt.Errorf("invalid yaml: %w", err)
		}

		return []byte(y), nil
	case []byte:
		err := yaml.Unmarshal(y, &out)
		if err != nil {
			return nil, fmt.Errorf("invalid yaml: %w", err)
		}

		return y, nil
	default:
		data, err := yaml.MarshalWithOptions(input, yamlEncodeOptions...)
		if err != nil {
			return nil, fmt.Errorf("invalid yaml: %w", err)
		}

		return data, nil
	}
}

func applyYAMLMatchers(b []byte, matchers ...match.YAMLMatcher) ([]byte, []match.MatcherError) {
	errors := []match.MatcherError{}

	for _, m := range matchers {
		y, errs := m.YAML(b)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}

		b = y
	}

	return b, errors
}

func takeYAMLSnapshot(b []byte) string {
	return escapeEndChars(string(b))
}
