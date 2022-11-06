package snaps

import (
	"errors"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
	"github.com/gkampitakis/go-snaps/match"
)

const jsonFilename = "matchJSON_test.snap"

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
				snapPath := setupSnapshot(t, jsonFilename, false)

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

	t.Run("should validate json", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			input interface{}
			err   string
		}{
			{
				name:  "string",
				input: "",
				err:   "invalid json",
			},
			{
				name:  "byte",
				input: []byte(`{"user"`),
				err:   "invalid json",
			},
			{
				name:  "struct",
				input: make(chan struct{}),
				err:   "json: unsupported type: chan struct {}",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				setupSnapshot(t, jsonFilename, false)

				mockT := test.MockTestingT{
					MockHelper: func() {},
					MockName: func() string {
						return "mock-name"
					},
					MockError: func(args ...interface{}) {
						test.Equal(t, tc.err, (args[0].(error)).Error())
					},
					MockLog: func(args ...interface{}) {
						// this is called when snapshot is written successfully
						test.NotCalled(t)
					},
				}

				MatchJSON(mockT, tc.input)
			})
		}
	})

	t.Run("matchers", func(t *testing.T) {
		t.Run("should apply matchers in order", func(t *testing.T) {
			setupSnapshot(t, jsonFilename, false)

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

			c1 := func(val interface{}) (interface{}, error) {
				return map[string]interface{}{"key2": nil}, nil
			}
			c2 := func(val interface{}) (interface{}, error) {
				return map[string]interface{}{"key3": nil}, nil
			}
			c3 := func(val interface{}) (interface{}, error) {
				return map[string]interface{}{"key4": nil}, nil
			}

			MatchJSON(
				mockT,
				`{"key1":""}`,
				match.Custom("key1", c1),
				match.Custom("key1.key2", c2),
				match.Custom("key1.key2.key3", c3),
			)
		})

		t.Run("should aggregate errors from matchers", func(t *testing.T) {
			setupSnapshot(t, jsonFilename, false)

			mockT := test.MockTestingT{
				MockHelper: func() {},
				MockName: func() string {
					return "mock-name"
				},
				MockError: func(args ...interface{}) {
					test.Equal(t,
						"\x1b[31;1m\n✕ match.Custom(\"age\") - mock error"+
							"\x1b[0m\x1b[31;1m\n✕ match.Any(\"missing.key.1\") - path does not exist"+
							"\x1b[0m\x1b[31;1m\n✕ match.Any(\"missing.key.2\") - path does not exist\x1b[0m",
						args[0],
					)
				},
				MockLog: func(args ...interface{}) { test.NotCalled(t) },
			}

			c := func(val interface{}) (interface{}, error) {
				return nil, errors.New("mock error")
			}
			MatchJSON(
				mockT,
				`{"age":10}`,
				match.Custom("age", c),
				match.Any("missing.key.1", "missing.key.2"),
			)
		})
	})

	t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {
		snapPath := setupSnapshot(t, jsonFilename, true)

		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				test.Equal(t, errSnapNotFound, args[0])
			},
			MockLog: func(args ...interface{}) {
				test.NotCalled(t)
			},
		}

		MatchJSON(mockT, "{}")

		_, err := snapshotFileToString(snapPath)
		test.Equal(t, errSnapNotFound, err)
		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		snapPath := setupSnapshot(t, jsonFilename, false, true)

		printerExpectedCalls := []func(received interface{}){
			func(received interface{}) { test.Equal(t, addedMsg, received) },
			func(received interface{}) { test.Equal(t, updatedMsg, received) },
		}
		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockName: func() string {
				return "mock-name"
			},
			MockError: func(args ...interface{}) {
				test.NotCalled(t)
			},
			MockLog: func(args ...interface{}) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			},
		}

		// First call for creating the snapshot
		MatchJSON(mockT, "{\"value\":\"hello world\"}")
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchJSON(mockT, "{\"value\":\"bye world\"}")

		snap, err := snapshotFileToString(snapPath)
		test.Nil(t, err)
		test.Equal(t, "\n[mock-name - 1]\n{\n \"value\": \"bye world\"\n}\n---\n", snap)
		test.Equal(t, 1, testEvents.items[updated])
	})
}
