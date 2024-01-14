package examples

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	v := m.Run()

	snaps.Clean(m)

	os.Exit(v)
}

func TestMatchSnapshot(t *testing.T) {
	t.Run("should make an int snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, 5)
	})

	t.Run("should make a string snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, "string snapshot")
	})

	t.Run("should make a map snapshot", func(t *testing.T) {
		m := map[string]any{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": func() {},
			"mock-3": float32(10.4),
		}

		snaps.MatchSnapshot(t, m)
	})

	t.Run("should make multiple entries in snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, 5, 10, 20, 25)
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

	t.Run("withConfig", func(t *testing.T) {
		t.Run("should allow changing filename", func(t *testing.T) {
			snaps.WithConfig(
				snaps.Filename("custom_file"),
			).MatchSnapshot(t, "snapshot data")
		})

		t.Run("should allow changing dir", func(t *testing.T) {
			s := snaps.WithConfig(snaps.Dir("testdata"))
			s.MatchSnapshot(t, "snapshot with different dir name")
			s.MatchSnapshot(t, "another one", 1, 10)
		})

		t.Run("should allow absolute path", func(t *testing.T) {
			_, b, _, _ := runtime.Caller(0)
			basepath := filepath.Dir(b)

			snaps.WithConfig(snaps.Dir(basepath+"/absolute_path")).
				MatchSnapshot(t, "supporting absolute path")
		})

		s := snaps.WithConfig(snaps.Dir("special_data"), snaps.Filename("different_name"))
		s.MatchSnapshot(t, "different data than the rest")
		snaps.MatchSnapshot(t, "this should use the default config")
	})
}

func TestSimpleTable(t *testing.T) {
	type testCases struct {
		description string
		input       any
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
			input: map[string]any{
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
