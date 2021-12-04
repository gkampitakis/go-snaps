package examples

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestParallel(t *testing.T) {
	type testCases struct {
		description string
		input       interface{}
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
