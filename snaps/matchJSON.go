package snaps

import (
	"encoding/json"
	"errors"

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

matchers is placeholder for now.
*/
func MatchJSON(t testingT, input interface{}, matchers ...interface{}) {
	t.Helper()

	dir, snapPath := snapDirAndName()
	testID := testsRegistry.getTestID(t.Name(), snapPath)

	j, err := validateJSON(input)
	if err != nil {
		handleError(t, err)
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

	diff := prettyDiff(prevSnapshot, snapshot)
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
