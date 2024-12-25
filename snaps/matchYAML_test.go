package snaps

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
	"github.com/gkampitakis/go-snaps/match"
)

const yamlFilename = "matchYAML_test.snap"

func TestMatchYAML(t *testing.T) {
	t.Run("should create yaml snapshot", func(t *testing.T) {
		expected := `user: mock-name
items:
  - 5
  - 1
  - 3
  - 4
`

		for _, tc := range []struct {
			name  string
			input any
		}{
			{
				name:  "string",
				input: "user: mock-name\nitems:\n  - 5\n  - 1\n  - 3\n  - 4\n",
			},
			{
				name:  "byte",
				input: []byte("user: mock-name\nitems:\n  - 5\n  - 1\n  - 3\n  - 4\n"),
			},
			{
				name: "marshal object",
				input: struct {
					User  string `yaml:"user"`
					Items []int  `yaml:"items"`
				}{
					User:  "mock-name",
					Items: []int{5, 1, 3, 4},
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				snapPath := setupSnapshot(t, yamlFilename, false)

				mockT := test.NewMockTestingT(t)
				mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

				MatchYAML(mockT, tc.input)

				snap, line, err := getPrevSnapshot("[mock-name - 1]", snapPath)

				test.NoError(t, err)
				test.Equal(t, 2, line)
				test.Equal(t, expected, snap)
				test.Equal(t, 1, testEvents.items[added])
				// clean up function called
				test.Equal(t, 0, testsRegistry.running[snapPath]["mock-name"])
				test.Equal(t, 1, testsRegistry.cleanup[snapPath]["mock-name"])
			})
		}

		t.Run("should validate yaml", func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				input any
				err   string
			}{
				{
					name:  "string",
					input: "key1: \"value1\nkey2: \"value2\"",
					err: `invalid yaml: [2:8] value is not allowed in this context. map key-value is pre-defined
   1 | key1: "value1
>  2 | key2: "value2"
              ^
`,
				},
				{
					name:  "byte",
					input: []byte("key1: \"value1\nkey2: \"value2\""),
					err: `invalid yaml: [2:8] value is not allowed in this context. map key-value is pre-defined
   1 | key1: "value1
>  2 | key2: "value2"
              ^
`,
				},
				{
					name:  "struct",
					input: make(chan struct{}),
					err:   "invalid yaml: unknown value type chan struct {}",
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					setupSnapshot(t, yamlFilename, false)

					mockT := test.NewMockTestingT(t)
					mockT.MockError = func(args ...any) {
						test.Equal(t, tc.err, (args[0].(error)).Error())
					}

					MatchYAML(mockT, tc.input)
				})
			}
		})

		t.Run("matchers", func(t *testing.T) {
			t.Run("should apply matches in order", func(t *testing.T) {
				setupSnapshot(t, yamlFilename, false)

				mockT := test.NewMockTestingT(t)
				mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }
				mockT.MockError = func(a ...any) {
					fmt.Println(a)
				}

				c1 := func(val any) (any, error) {
					return map[string]any{"key2": nil}, nil
				}
				c2 := func(val any) (any, error) {
					return map[string]any{"key3": nil}, nil
				}
				c3 := func(val any) (any, error) {
					return map[string]any{"key4": nil}, nil
				}

				MatchYAML(mockT, "key1: \"\"",
					match.Custom("$.key1", c1),
					match.Custom("$.key1.key2", c2),
					match.Custom("$.key1.key2.key3", c3),
				)
			})

			t.Run("should aggregate errors from matchers", func(t *testing.T) {
				setupSnapshot(t, yamlFilename, false)

				mockT := test.NewMockTestingT(t)
				mockT.MockError = func(args ...any) {
					test.Equal(t,
						"\x1b[31;1m\n✕ match.Custom(\"$.age\") - mock error"+
							"\x1b[0m\x1b[31;1m\n✕ match.Any(\"$.missing.key.1\") - path does not exist"+
							"\x1b[0m\x1b[31;1m\n✕ match.Any(\"$.missing.key.2\") - path does not exist\x1b[0m",
						args[0],
					)
				}

				c := func(val any) (any, error) {
					return nil, errors.New("mock error")
				}
				MatchYAML(
					mockT,
					`age: 10`,
					match.Custom("$.age", c),
					match.Any("$.missing.key.1", "$.missing.key.2"),
				)
			})
		})

		t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {
			setupSnapshot(t, yamlFilename, true)

			mockT := test.NewMockTestingT(t)
			mockT.MockError = func(args ...any) {
				test.Equal(t, errSnapNotFound, args[0].(error))
			}

			MatchYAML(mockT, "")

			test.Equal(t, 1, testEvents.items[erred])
		})

		t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
			snapPath := setupSnapshot(t, yamlFilename, false, true)
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
			MatchYAML(mockT, "value: hello world")
			test.Equal(t, 1, testEvents.items[added])

			// Resetting registry to emulate the same MatchSnapshot call
			testsRegistry = newRegistry()

			// Second call with different params
			MatchYAML(mockT, "value: bye world")

			test.Equal(
				t,
				"\n[mock-name - 1]\nvalue: bye world\n---\n",
				test.GetFileContent(t, snapPath),
			)
			test.Equal(t, 1, testEvents.items[updated])
		})
	})
}
