package snaps

import (
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

const standaloneFilename = "matchStandaloneSnapshot_test_mock-name_1.snap"

func TestMatchStandaloneSnapshot(t *testing.T) {
	t.Run("should create snapshot", func(t *testing.T) {
		snapPath := setupSnapshot(t, standaloneFilename, false)
		mockT := test.NewMockTestingT(t)
		mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

		MatchStandaloneSnapshot(mockT, "hello world")

		test.Equal(t, "hello world", test.GetFileContent(t, snapPath))
		test.Equal(t, 1, testEvents.items[added])
		// clean up function called

		registryKey := filepath.Join(
			filepath.Dir(snapPath),
			"matchStandaloneSnapshot_test_mock-name_%d.snap",
		)
		test.Equal(t, 0, standaloneTestsRegistry.running[registryKey])
		test.Equal(t, 1, standaloneTestsRegistry.cleanup[registryKey])
	})

	t.Run("should pass tests with no diff", func(t *testing.T) {
		snapPath := setupSnapshot(t, standaloneFilename, false, false)

		printerExpectedCalls := []func(received any){
			func(received any) { test.Equal(t, addedMsg, received.(string)) },
			func(received any) { t.Error("should not be called 3rd time") },
		}
		mockT := test.NewMockTestingT(t)
		mockT.MockLog = func(args ...any) {
			printerExpectedCalls[0](args[0])

			// shift
			printerExpectedCalls = printerExpectedCalls[1:]
		}

		s := WithConfig(Update(true))
		// First call for creating the snapshot
		s.MatchStandaloneSnapshot(mockT, "hello world")
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchStandaloneSnapshot call
		standaloneTestsRegistry = newStandaloneRegistry()

		// Second call with same params
		s.MatchStandaloneSnapshot(mockT, "hello world")

		test.Equal(t, "hello world", test.GetFileContent(t, snapPath))
		test.Equal(t, 1, testEvents.items[passed])
	})

	t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {
		setupSnapshot(t, standaloneFilename, true)

		mockT := test.NewMockTestingT(t)
		mockT.MockError = func(args ...any) {
			test.Equal(t, errSnapNotFound, args[0].(error))
		}

		MatchStandaloneSnapshot(mockT, "hello world")

		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should return error when diff is found", func(t *testing.T) {
		setupSnapshot(t, standaloneFilename, false)

		printerExpectedCalls := []func(received any){
			func(received any) { test.Equal(t, addedMsg, received.(string)) },
			func(received any) { t.Error("should not be called 2nd time") },
		}
		mockT := test.NewMockTestingT(t)
		mockT.MockError = func(args ...any) {
			expected := "\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 1\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m" +
				"+ Received + 1\x1b[0m\n\n\x1b[48;5;225m\x1b[38;5;52m- \x1b[0m\x1b[48;5;127m\x1b[38;5;255m" +
				"hello\x1b[0m\x1b[48;5;225m\x1b[38;5;52m world\x1b[0m\n\x1b[48;5;159m\x1b[38;5;22m" +
				"+ \x1b[0m\x1b[48;5;23m\x1b[38;5;255mbye\x1b[0m\x1b[48;5;159m\x1b[38;5;22m world\x1b[0m\n\n\x1b[2m" +
				"at " + filepath.FromSlash(
				"__snapshots__/matchStandaloneSnapshot_test_mock-name_1.snap:1",
			) + "\n\x1b[0m"

			test.Equal(t, expected, args[0].(string))
		}
		mockT.MockLog = func(args ...any) {
			printerExpectedCalls[0](args[0])

			// shift
			printerExpectedCalls = printerExpectedCalls[1:]
		}

		// First call for creating the snapshot
		MatchStandaloneSnapshot(mockT, "hello world")
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchStandaloneSnapshot call
		standaloneTestsRegistry = newStandaloneRegistry()

		// Second call with different data
		MatchStandaloneSnapshot(mockT, "bye world")
		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should update snapshot", func(t *testing.T) {
		t.Run("when 'updateVAR==true'", func(t *testing.T) {
			snapPath := setupSnapshot(t, standaloneFilename, false, true)

			printerExpectedCalls := []func(received any){
				func(received any) { test.Equal(t, addedMsg, received.(string)) },
				func(received any) { test.Equal(t, updatedMsg, received.(string)) },
				func(received any) { t.Error("should not be called 3rd time") },
			}
			mockT := test.NewMockTestingT(t)
			mockT.MockLog = func(args ...any) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			}

			// First call for creating the snapshot
			MatchStandaloneSnapshot(mockT, "hello world")
			test.Equal(t, 1, testEvents.items[added])

			// Resetting registry to emulate the same MatchStandaloneSnapshot call
			standaloneTestsRegistry = newStandaloneRegistry()

			// Second call with different params
			MatchStandaloneSnapshot(mockT, "bye world")

			test.Equal(t, "bye world", test.GetFileContent(t, snapPath))
			test.Equal(t, 1, testEvents.items[updated])
		})

		t.Run("when config update", func(t *testing.T) {
			snapPath := setupSnapshot(t, standaloneFilename, false, false)

			printerExpectedCalls := []func(received any){
				func(received any) { test.Equal(t, addedMsg, received.(string)) },
				func(received any) { test.Equal(t, updatedMsg, received.(string)) },
				func(received any) { t.Error("should not be called 3rd time") },
			}
			mockT := test.NewMockTestingT(t)
			mockT.MockLog = func(args ...any) {
				printerExpectedCalls[0](args[0])

				// shift
				printerExpectedCalls = printerExpectedCalls[1:]
			}

			s := WithConfig(Update(true))
			// First call for creating the snapshot
			s.MatchStandaloneSnapshot(mockT, "hello world")
			test.Equal(t, 1, testEvents.items[added])

			// Resetting registry to emulate the same MatchStandaloneSnapshot call
			standaloneTestsRegistry = newStandaloneRegistry()

			// Second call with different params
			s.MatchStandaloneSnapshot(mockT, "bye world")

			test.Equal(t, "bye world", test.GetFileContent(t, snapPath))
			test.Equal(t, 1, testEvents.items[updated])
		})
	})
}
