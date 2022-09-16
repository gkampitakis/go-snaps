package examples

import (
	"bytes"
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(t *testing.M) {
	v := t.Run()

	snaps.Clean(t)

	os.Exit(v)
}

func TestSimple(t *testing.T) {
	t.Run("should make an int snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, 5)
	})

	t.Run("should make a string snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, "string snapshot")
	})

	t.Run("should make a map snapshot", func(t *testing.T) {
		m := map[string]interface{}{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": func() {},
			"mock-3": float32(10.4),
		}

		snaps.MatchSnapshot(t, m)
	})

	t.Run("should make create multiple snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, 1000)
		snaps.MatchSnapshot(t, "another snapshot")
		snaps.MatchSnapshot(t, `{
			"user": "gkampitakis",
			"id": 1234567,
			"data": [ ]
		}`)
	})

	t.Run("nest", func(t *testing.T) {
		t.Run("more", func(t *testing.T) {
			t.Run("one more nested test", func(t *testing.T) {
				snaps.MatchSnapshot(t, "it's okay")
			})
		})
	})

	t.Run(".*", func(t *testing.T) {
		snaps.MatchSnapshot(t, "ignore regex patterns on names")
	})
}

func TestSimpleTable(t *testing.T) {
	type testCases struct {
		description string
		input       interface{}
	}

	for _, scenario := range []testCases{
		{
			description: "string",
			input:       "input",
		},
		{
			description: "integer",
			input:       10,
		},
		{
			description: "map",
			input: map[string]interface{}{
				"test": func() {},
			},
		},
		{
			description: "buffer",
			input:       bytes.NewBufferString("Buffer string"),
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			snaps.MatchSnapshot(t, scenario.input)
		})
	}
}
