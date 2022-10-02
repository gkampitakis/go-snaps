package snaps

import (
	"errors"
	"fmt"
	"log"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
)

// TODO: matchers need a type and some scoping
// they should accept
func MatchJSON(t testingT, input interface{}, matchers ...match.JSONMatcher) {
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
		return nil, fmt.Errorf("type: %T not supported", j)
	}
}

func takeJSONSnapshot(b []byte) string {
	return string(pretty.PrettyOptions(b, &pretty.Options{
		SortKeys: true,
		Indent:   " ",
	}))
}

func validateMatchers(b []byte, matchers ...match.JSONMatcher) []byte {
	for _, m := range matchers {
		bb, errString := m(b)
		if errString != "" {
			log.Println(errString)
		}
		b = bb
	}
	return b
}
