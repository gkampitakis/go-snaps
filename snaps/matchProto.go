package snaps

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	errInvalidProto = errors.New("invalid proto message")
	errNilProto     = errors.New("proto message cannot be nil")
)

/*
MatchProto verifies the input protobuf message matches the most recent snap file.
Input must be a proto.Message.

	snaps.MatchProto(t, &pb.User{Name: "mock-user", Age: 10})

The protobuf data is saved in the snapshot file using protojson. When comparing,
the snapshot is unmarshaled back into the proto type and compared using
cmp.Diff with protocmp.Transform().
*/
func (c *Config) MatchProto(t testingT, input proto.Message) {
	t.Helper()

	matchProto(c, t, input)
}

/*
MatchProto verifies the input protobuf message matches the most recent snap file.
Input must be a proto.Message.

	snaps.MatchProto(t, &pb.User{Name: "mock-user", Age: 10})

The protobuf data is saved in the snapshot file using protojson. When comparing,
the snapshot is unmarshaled back into the proto type and compared using
cmp.Diff with protocmp.Transform().
*/
func MatchProto(t testingT, input proto.Message) {
	t.Helper()

	matchProto(&defaultConfig, t, input)
}

func matchProto(c *Config, t testingT, input proto.Message) {
	t.Helper()

	snapPath, snapPathRel := snapshotPath(c, t.Name(), false)
	testID := testsRegistry.getTestID(snapPath, t.Name())
	t.Cleanup(func() {
		testsRegistry.reset(snapPath, t.Name())
	})

	if input == nil {
		handleError(t, errNilProto)
		return
	}

	// Marshal proto to JSON for storage with consistent formatting
	opts := protojson.MarshalOptions{
		Multiline: true,
		Indent:    " ",
		EmitUnpopulated: false,
	}
	protoJSON, err := opts.Marshal(input)
	if err != nil {
		handleError(t, fmt.Errorf("%w: %v", errInvalidProto, err))
		return
	}

	snapshot := takeProtoSnapshot(protoJSON)
	prevSnapshot, line, err := getPrevSnapshot(testID, snapPath)
	if errors.Is(err, errSnapNotFound) {
		if !shouldCreate(c.update) {
			handleError(t, err)
			return
		}

		err := addNewSnapshot(testID, snapshot, snapPath)
		if err != nil {
			handleError(t, err)
			return
		}

		t.Log(addedMsg)
		testEvents.register(added)
		return
	}
	if err != nil {
		handleError(t, err)
		return
	}

	// Compare protos using cmp.Diff with protocmp.Transform()
	diff := compareProtos(c, t, input, prevSnapshot, snapPathRel, line)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(c.update) {
		handleError(t, diff)
		return
	}

	if err = updateSnapshot(testID, snapshot, snapPath); err != nil {
		handleError(t, err)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}

func takeProtoSnapshot(jsonData []byte) string {
	return strings.TrimSuffix(string(jsonData), "\n")
}


func compareProtos(c *Config, t testingT, current proto.Message, prevSnapshotJSON string, snapPathRel string, line int) string {
	t.Helper()

	// Create a new instance of the same proto type
	protoType := reflect.TypeOf(current)
	if protoType.Kind() == reflect.Ptr {
		protoType = protoType.Elem()
	}

	newProto := reflect.New(protoType).Interface().(proto.Message)

	// Unmarshal the previous snapshot JSON back into the proto type
	err := protojson.Unmarshal([]byte(prevSnapshotJSON), newProto)
	if err != nil {
		// If unmarshaling fails, fall back to string comparison
		currentJSON, marshalErr := protojson.MarshalOptions{
			Multiline: true,
			Indent:    " ",
		}.Marshal(current)
		if marshalErr != nil {
			return fmt.Sprintf("failed to unmarshal previous snapshot: %v\nfailed to marshal current proto: %v", err, marshalErr)
		}
		return prettyDiff(prevSnapshotJSON, string(currentJSON), snapPathRel, line)
	}

	// Use cmp.Diff with protocmp.Transform() to compare
	diff := cmp.Diff(
		newProto,
		current,
		protocmp.Transform(),
	)

	if diff == "" {
		return ""
	}

	// Format the diff similar to how other match functions do
	return formatProtoDiff(diff, snapPathRel, line)
}

func formatProtoDiff(diff, snapPathRel string, line int) string {
	if diff == "" {
		return ""
	}

	// Count inserted and deleted lines for the header
	// cmp.Diff output format: lines with "+" or "-" after optional whitespace
	lines := strings.Split(strings.TrimSuffix(diff, "\n"), "\n")
	var inserted, deleted int
	for _, l := range lines {
		trimmed := strings.TrimLeft(l, " \t")
		if strings.HasPrefix(trimmed, "+") {
			inserted++
		} else if strings.HasPrefix(trimmed, "-") {
			deleted++
		}
	}

	var s strings.Builder
	s.Grow(len(diff) + 100)

	// Build header similar to buildDiffReport
	iPadding, dPadding := intPadding(inserted, deleted)

	s.WriteByte('\n')
	colors.FprintDelete(&s, fmt.Sprintf("Snapshot %s- %d\n", dPadding, deleted))
	colors.FprintInsert(&s, fmt.Sprintf("Received %s+ %d\n", iPadding, inserted))
	s.WriteByte('\n')

	// Format diff lines with colors
	// cmp.Diff uses indentation, so we check for "-" and "+" after trimming whitespace
	for _, l := range lines {
		trimmed := strings.TrimLeft(l, " \t")
		if strings.HasPrefix(trimmed, "+") {
			colors.FprintInsert(&s, l+"\n")
		} else if strings.HasPrefix(trimmed, "-") {
			colors.FprintDelete(&s, l+"\n")
		} else {
			colors.FprintEqual(&s, l+"\n")
		}
	}

	s.WriteByte('\n')
	if snapPathRel != "" {
		colors.Fprint(&s, colors.Dim, fmt.Sprintf("at %s:%d\n", snapPathRel, line))
	}

	return s.String()
}
