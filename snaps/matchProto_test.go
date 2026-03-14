package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

const protoFilename = "matchProto_test.snap"

// TestMatchProto tests the MatchProto function
// Note: This test uses structpb.Value as a simple proto message for testing
func TestMatchProto(t *testing.T) {
	t.Run("should create proto snapshot", func(t *testing.T) {
		snapPath := setupSnapshot(t, protoFilename, false)

		mockT := test.NewMockTestingT(t)
		mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

		// Create a simple proto message using structpb
		protoMsg := structpb.NewStringValue("mock-value")

		MatchProto(mockT, protoMsg)

		snap, line, err := getPrevSnapshot("[mock-name - 1]", snapPath)

		test.NoError(t, err)
		test.Equal(t, 2, line)
		// The snapshot should contain JSON representation of the proto
		test.Contains(t, snap, "mock-value")
		test.Equal(t, 1, testEvents.items[added])
		// clean up function called
		test.Equal(t, 0, testsRegistry.running[snapPath]["mock-name"])
		test.Equal(t, 1, testsRegistry.cleanup[snapPath]["mock-name"])
	})

	t.Run("should validate nil proto", func(t *testing.T) {
		setupSnapshot(t, protoFilename, false)

		mockT := test.NewMockTestingT(t)
		mockT.MockError = func(args ...any) {
			test.Equal(t, errNilProto, args[0].(error))
		}

		MatchProto(mockT, nil)

		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("if it's running on ci should skip creating snapshot", func(t *testing.T) {
		setupSnapshot(t, protoFilename, true)

		mockT := test.NewMockTestingT(t)
		mockT.MockError = func(args ...any) {
			test.Equal(t, errSnapNotFound, args[0].(error))
		}

		protoMsg := structpb.NewStringValue("test")
		MatchProto(mockT, protoMsg)

		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("if snaps.Update(false) should skip creating snapshot", func(t *testing.T) {
		setupSnapshot(t, protoFilename, false)

		mockT := test.NewMockTestingT(t)
		mockT.MockError = func(args ...any) {
			test.Equal(t, errSnapNotFound, args[0].(error))
		}

		protoMsg := structpb.NewStringValue("test")
		WithConfig(Update(false)).MatchProto(mockT, protoMsg)

		test.Equal(t, 1, testEvents.items[erred])
	})

	t.Run("should update snapshot when 'shouldUpdate'", func(t *testing.T) {
		snapPath := setupSnapshot(t, protoFilename, false, "true")
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
		protoMsg1 := structpb.NewStringValue("hello world")
		MatchProto(mockT, protoMsg1)
		test.Equal(t, 1, testEvents.items[added])

		// Resetting registry to emulate the same MatchProto call
		testsRegistry = newRegistry()

		// Second call with different params
		protoMsg2 := structpb.NewStringValue("bye world")
		MatchProto(mockT, protoMsg2)

		test.Contains(t, test.GetFileContent(t, snapPath), "bye world")
		test.Equal(t, 1, testEvents.items[updated])
	})

	t.Run(
		"should create and update snapshot when UPDATE_SNAPS=always even on CI",
		func(t *testing.T) {
			snapPath := setupSnapshot(t, protoFilename, true, "always")

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
			protoMsg1 := structpb.NewStringValue("hello world")
			WithConfig(Update(false)).MatchProto(mockT, protoMsg1)
			test.Equal(t, 1, testEvents.items[added])

			// Resetting registry to emulate the same MatchProto call
			testsRegistry = newRegistry()

			// Second call with different params
			protoMsg2 := structpb.NewStringValue("bye world")
			WithConfig(Update(false)).MatchProto(mockT, protoMsg2)

			test.Contains(t, test.GetFileContent(t, snapPath), "bye world")
			test.Equal(t, 1, testEvents.items[updated])
		},
	)

	t.Run("should accept protocmp options", func(t *testing.T) {
		setupSnapshot(t, protoFilename, false)

		mockT := test.NewMockTestingT(t)
		mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

		protoMsg := structpb.NewStringValue("test")

		// Should work with protocmp options (even if they don't do anything useful here)
		// The options are passed through to cmp.Diff
		MatchProto(mockT, protoMsg, protocmp.Transform(), cmp.AllowUnexported())
		test.Equal(t, 1, testEvents.items[added])
	})

	t.Run("should work with multiple protocmp options", func(t *testing.T) {
		snapPath := setupSnapshot(t, protoFilename, false)

		mockT := test.NewMockTestingT(t)
		mockT.MockLog = func(args ...any) { test.Equal(t, addedMsg, args[0].(string)) }

		protoMsg := structpb.NewStringValue("test")

		// Should work with multiple options
		MatchProto(mockT, protoMsg, protocmp.Transform(), cmp.AllowUnexported())
		test.Equal(t, 1, testEvents.items[added])

		snap, _, err := getPrevSnapshot("[mock-name - 1]", snapPath)
		test.NoError(t, err)
		test.Contains(t, snap, "test")
	})
}
