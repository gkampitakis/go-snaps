package snaps

import (
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"github.com/tidwall/sjson"
)

// TODO: matchers need a type and some scoping
// they should accept
func MatchJSON(t testingT, input interface{}, matchers ...func([]byte) []byte) {
	t.Helper()

	j, err := validateJSON(input)
	if err != nil {
		handleError(t, err)
		return
	}

	if len(matchers) > 0 {
		// here the matchers are validate
		// and the json input has all the values replaced with placeholders
		// and it's ready to be passed down for string diffing
		j = validateMatchers(j, matchers...)
	}

	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)
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

// TODO: we can write our own validator here maybe
// and put more meaningful errors
func validateJSON(input interface{}) ([]byte, error) {
	switch j := input.(type) {
	case string:
		if !gjson.Valid(j) {
			return nil, errors.New("invalid json")
		}

		return []byte(j), nil
	case []byte:
		if !gjson.ValidBytes(j) {
			return nil, errors.New("invalid json")
		}

		return j, nil
	default:
		return nil, errors.New(fmt.Sprintf("type: %T not supported\n", j))
	}
}

func takeJSONSnapshot(b []byte) string {
	return string(pretty.PrettyOptions(b, &pretty.Options{
		SortKeys: true,
		Indent:   " ",
	}))
}

func Ignore(pattern ...string) func([]byte) []byte {
	return func(s []byte) []byte {
		newJSON := s
		for _, p := range pattern {
			newJSON, _ = sjson.SetBytes(newJSON, p, "<ignore value>")
		}
		return newJSON
	}
}

func validateMatchers(b []byte, matchers ...func([]byte) []byte) []byte {
	for _, m := range matchers {
		b = m(b)
	}
	return b
}
