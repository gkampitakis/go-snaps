package snaps

import (
	"testing"

	"github.com/kr/pretty"
)

func TestDiff(t *testing.T) {
	t.Run("should return empty string if no diffs", func(t *testing.T) {
		expected, received := "Hello World\n", "Hello World\n"

		if diff := prettyDiff(expected, received); diff != "" {
			t.Errorf("found diff between same string %s", diff)
		}
	})

	t.Run("should return red diff", func(t *testing.T) {
		expected, received := "Hello", ""
		expectedDiff := "\n\x1b[41m\x1b[37;1m Snapshot \x1b[0m\n\x1b[42m\x1b[37;1m Received \x1b[0m\n\n\x1b[31;1mHello\x1b[0m\n"

		if diff := prettyDiff(expected, received); diff != expectedDiff {
			t.Errorf("found diff between same string %s", diff)
		}
	})

	t.Run("should return green diff", func(t *testing.T) {
		expected, received := "", "Hello"
		expectedDiff := "\n\x1b[41m\x1b[37;1m Snapshot \x1b[0m\n\x1b[42m\x1b[37;1m Received \x1b[0m\n\n\x1b[32;1mHello\x1b[0m\n"

		if diff := prettyDiff(expected, received); diff != expectedDiff {
			t.Errorf("found diff between same string %s", diff)
		}
	})

	t.Run("should return green bg diff for spaces", func(t *testing.T) {
		expected, received := "    ", ""
		expectedDiff := "\n\x1b[41m\x1b[37;1m Snapshot \x1b[0m\n\x1b[42m\x1b[37;1m Received \x1b[0m\n\n\x1b[41m\x1b[37;1m    \x1b[0m\n"

		if diff := prettyDiff(expected, received); diff != expectedDiff {
			t.Errorf("found diff between same string %s", diff)
		}
	})

	t.Run("should return red bg diff for spaces", func(t *testing.T) {
		expected, received := "", "    "
		expectedDiff := "\n\x1b[41m\x1b[37;1m Snapshot \x1b[0m\n\x1b[42m\x1b[37;1m Received \x1b[0m\n\n\x1b[42m\x1b[37;1m    \x1b[0m\n"

		if diff := prettyDiff(expected, received); diff != expectedDiff {
			t.Errorf("found diff between same string %s", diff)
		}
	})

	t.Run("should colorize space diffs", func(t *testing.T) {
		expected := `{
			"user": "gkampitakis",
			"id": 1234567,
			"data": [ ]
		}`
		received := `{
					"user": "gk",
					"id": 1234567,
					"data": [ ]
		}`

		expectedDiff := "\n\x1b[41m\x1b[37;1m Snapshot \x1b[0m\n\x1b[42m\x1b[37;1m " +
			"Received \x1b[0m\n\n\x1b[2m{\n            \x1b[0m\x1b[42m\x1b[37;1m        \x1b[" +
			"0m\x1b[2m\"user\": \"gk\x1b[0m\x1b[31;1mampitakis\x1b[0m\x1b[2m\",\n            \x1b[0" +
			"m\x1b[42m\x1b[37;1m        \x1b[0m\x1b[2m\"id\": 1234567,\n\x1b[0m\x1b[42m\x1b[37;1m        \x1b[" +
			"0m\x1b[2m            \"data\": [ ]\n        }\x1b[0m\n"

		if diff := prettyDiff(pretty.Sprint(expected), pretty.Sprint(received)); diff != expectedDiff {
			t.Errorf("wrong diff produced %s", diff)
		}
	})
}
