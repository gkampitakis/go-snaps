package snaps

import (
	"errors"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
	"github.com/gkampitakis/go-snaps/match"
)

const jsonStandaloneFilename = "mock-name_1.snap.json"

func TestMatchStandaloneJSON(t *testing.T) {
	t.Run("should create json snapshot", func(t *testing.T) {
		expected := `{
 "items": [
  5,
  1,
  3,
  4
 ],
 "user": "mock-name"
}`

		for _, tc := range []struct {
			name  string
			input any
		}{
			{
				name:  "string",
				input: `{"user":"mock-name","items":[5,1,3,4]}`,
			},
			{
				name:  "byte",
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
				snapPath := setupSnapshot(t, jsonStandaloneFilename, false)

				mockT := test.NewMockTestingT(t)
				mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

				MatchStandaloneJSON(mockT, tc.input)

				snap := test.GetFileContent(t, snapPath)

				test.Equal(t, expected, snap)
				test.Equal(t, 1, testEvents.items[added])
				// clean up function called
				for _, v := range standaloneTestsRegistry.running {
					test.Equal(t, 0, v)
				}
				for _, v := range standaloneTestsRegistry.cleanup {
					test.Equal(t, 1, v)
				}
			})
		}
	})

	t.Run("should validate json", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			input any
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
				setupSnapshot(t, jsonStandaloneFilename, false)

				mockT := test.NewMockTestingT(t)
				mockT.MockError = func(args ...any) {
					test.Equal(t, tc.err, (args[0].(error)).Error())
				}

				MatchStandaloneJSON(mockT, tc.input)
			})
		}
	})

	t.Run("matchers", func(t *testing.T) {
		t.Run("should apply matchers in order", func(t *testing.T) {
			snapPath := setupSnapshot(t, jsonStandaloneFilename, false)

			mockT := test.NewMockTestingT(t)
			mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

			c1 := func(val any) (any, error) {
				return map[string]any{"key2": nil}, nil
			}
			c2 := func(val any) (any, error) {
				return map[string]any{"key3": nil}, nil
			}
			c3 := func(val any) (any, error) {
				return map[string]any{"key4": nil}, nil
			}

			MatchStandaloneJSON(
				mockT,
				`{"key1":""}`,
				match.Custom("key1", c1),
				match.Custom("key1.key2", c2),
				match.Custom("key1.key2.key3", c3),
			)

			test.Equal(
				t,
				"{\n \"key1\": {\n  \"key2\": {\n   \"key3\": {\n    \"key4\": null\n   }\n  }\n }\n}",
				test.GetFileContent(t, snapPath),
			)
		})

		t.Run("should aggregate errors from matchers", func(t *testing.T) {
			setupSnapshot(t, jsonStandaloneFilename, false)

			mockT := test.NewMockTestingT(t)
			mockT.MockError = func(args ...any) {
				test.Equal(t,
					"\x1b[31;1m\n✕ match.Custom(\"age\") - mock error"+
						"\x1b[0m\x1b[31;1m\n✕ match.Any(\"missing.key.1\") - path does not exist"+
						"\x1b[0m\x1b[31;1m\n✕ match.Any(\"missing.key.2\") - path does not exist\x1b[0m",
					args[0],
				)
			}

			c := func(val any) (any, error) {
				return nil, errors.New("mock error")
			}
			MatchStandaloneJSON(
				mockT,
				`{"age":10}`,
				match.Custom("age", c),
				match.Any("missing.key.1", "missing.key.2"),
			)
		})
	})

	t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {
		setupSnapshot(t, jsonStandaloneFilename, true)

		mockT := test.NewMockTestingT(t)
		mockT.MockError = func(args ...any) {
			test.Equal(t, errSnapNotFound, args[0].(error))
		}

		MatchStandaloneJSON(mockT, "{}")

		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		snapPath := setupSnapshot(t, jsonStandaloneFilename, false, true)

		printerExpectedCalls := []func(received any){
			func(received any) { test.Equal(t, addedMsg, received.(string)) },
			func(received any) { test.Equal(t, updatedMsg, received.(string)) },
		}
		mockT := test.NewMockTestingT(t)
		mockT.MockLog = func(args ...any) {
			printerExpectedCalls[0](args[0])

			// shift
			printerExpectedCalls = printerExpectedCalls[1:]
		}

		// First call for creating the snapshot
		MatchStandaloneJSON(mockT, "{\"value\":\"hello world\"}")
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchSnapshot call
		testsRegistry = newRegistry()

		// Second call with different params
		MatchStandaloneJSON(mockT, "{\"value\":\"bye world\"}")

		test.Equal(
			t,
			"{\n \"value\": \"bye world\"\n}",
			test.GetFileContent(t, snapPath),
		)
		test.Equal(t, 1, testEvents.items[updated])
	})
}
