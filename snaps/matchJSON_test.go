package snaps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/ciinfo"
	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestMatchJSON(t *testing.T) {
	t.Run("should create json snapshot", func(t *testing.T) {
		expected := `
{
 "items": [
  5,
  1,
  3,
  4
 ],
 "user": "mock-name"
}
`

		for _, tc := range []struct {
			name  string
			input interface{}
		}{
			{
				name:  "string",
				input: `{"user":"mock-name","items":[5,1,3,4]}`,
			},
			{
				name:  "string",
				input: []byte(`{"user":"mock-name","items":[5,1,3,4]}`),
			},
			{
				name: "marshal object",
				input: struct {
					User  string `json:"user"`
					Items []int  `json:"items"`
				}{
					User:  "mock-name",
					Items: []int{5, 1, 3, 4},
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				dir, _ := os.Getwd()
				snapPath := filepath.Join(dir, "__snapshots__", "matchJSON_test.snap")
				isCI = false

				t.Cleanup(func() {
					os.Remove(snapPath)
					testsRegistry = newRegistry()
					testEvents = newTestEvents()
					isCI = ciinfo.IsCI
				})

				mockT := test.MockTestingT{
					MockHelper: func() {},
					MockName: func() string {
						return "mock-name"
					},
					MockError: func(args ...interface{}) {
						test.NotCalled(t)
					},
					MockLog: func(args ...interface{}) { test.Equal(t, addedMsg, args[0]) },
				}

				MatchJSON(mockT, tc.input)

				snap, err := getPrevSnapshot("mock-name - 1", snapPath)
				test.Nil(t, err)
				test.Equal(t, expected, snap)
				test.Equal(t, 1, testEvents.items[added])
			})
		}
	})

	t.Run("should validate json", func(t *testing.T) {})

	t.Run("matchers", func(t *testing.T) {
		t.Run("should apply matchers in order", func(t *testing.T) {})

		t.Run("should aggregate errors from matchers", func(t *testing.T) {})
	})

	t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {})

	t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {})
}
