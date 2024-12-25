package examples

import (
	"bytes"
	"flag"
	"fmt"
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

// You can use -update flag to control if you want to update the snapshots
// go test ./... -v -update
var updateSnapshot = flag.Bool("update", false, "update snapshots flag")

func TestUpdateWithFlag(t *testing.T) {
	snaps := snaps.WithConfig(snaps.Update(*updateSnapshot))

	inputs := []string{
		"lore ipsum dolor sit amet",
		"consectetur adipiscing elit",
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
		"Ut enim ad minim veniam, quis nostrud laboris nisi ut aliquip ex ea commodo consequat.",
	}

	for i, input := range inputs {
		t.Run(fmt.Sprintf("test - %d", i), func(t *testing.T) {
			snaps.MatchSnapshot(t, input)
		})
	}
}

func TestParallel(t *testing.T) {
	type testCases struct {
		description string
		input       any
	}

	value := 10

	tests := []testCases{
		{
			description: "should snap an integer",
			input:       10,
		},
		{
			description: "should snap a float",
			input:       float64(10.5),
		},
		{
			description: "should snap a struct",
			input: struct {
				user  string
				email string
				age   int
			}{
				"gkampitakis",
				"mock@mail.com",
				10,
			},
		},
		{
			description: "should snap a struct with fields",
			input: struct {
				_    struct{}
				name string
				id   string
			}{
				name: "mock-name",
				id:   "123456",
			},
		},
		{
			description: "should snap an integer slice",
			input:       []int{1, 2, 3, 4},
		},
		{
			description: "should snap a map",
			input: map[string]int{
				"value-0": 0,
				"value-1": 1,
				"value-2": 2,
				"value-3": 3,
			},
		},
		{
			description: "should snap a buffer",
			input:       bytes.NewBufferString("Buffer string"),
		},
		{
			description: "should snap a pointer",
			input:       &value,
		},
	}

	for _, scenario := range tests {
		// capture range variable
		s := scenario

		t.Run(s.description, func(t *testing.T) {
			t.Parallel()

			snaps.MatchSnapshot(t, s.input)
		})
	}
}
