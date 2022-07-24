package snaps

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	diffEqual  diffmatchpatch.Operation = 0
	diffInsert diffmatchpatch.Operation = 1
	diffDelete diffmatchpatch.Operation = -1

	EqualOp  DiffOperation = 0
	InsertOp DiffOperation = 1
	DeleteOp DiffOperation = -1
	SkipOp   DiffOperation = 2
)

type DiffOperation int8

type processedDiff struct {
	Type   DiffOperation
	Items  []string // TODO: rename
	Concat bool
}

// shows 11 lines 5 not changed 1 changed 5 not changed
// @@ -53,11 +53,11 @@
// https://stackoverflow.com/a/6508925/10068782
// createPatchMark

/**

skip feature

if its first we need to keep the last 4 of the equal

if in middle we need 4 before and 4 after

if its last we need to keep the first 4 of the equal

*/

// TEST: how it looks with long snapshot more than 100 lines with small changes
// TODO: it's difficult to recognize diffs when deletes and additions are done on the same spot intermingled

func processDiffLines(diffs []diffmatchpatch.Diff) []processedDiff {
	processedDiffs := []processedDiff{}

	for i, diff := range diffs {
		text := diff.Text
		var lines []string
		if text == "\n" {
			lines = []string{text}
		} else {
			lines = strings.Split(text, newLine)
		}
		lastIdx := len(lines) - 1
		concat := true
		prevConcat := false

		if lines[lastIdx] == "" {
			lines = lines[:lastIdx]
			concat = false
		}

		if i > 0 {
			prevConcat = processedDiffs[i-1].Concat
		}

		switch diff.Type {
		case diffEqual:
			// we are in the start
			// TODO: clean this if else hell
			if len(lines) > 4 {
				// i == 0 &&
				// processedDiffs = append(processedDiffs, )

				skipStep := processedDiff{
					Type:   SkipOp,
					Items:  nil,
					Concat: false,
				}

				equalStep := processedDiff{
					Type:   DiffOperation(diffEqual),
					Items:  lines,
					Concat: prevConcat || concat,
				}

				if i == 0 {
					lines = lines[len(lines)-4:]
					equalStep.Items = lines

					processedDiffs = append(processedDiffs, skipStep, equalStep)
				} else if i+1 == len(diffs) {
					lines = lines[:4]
					equalStep.Items = lines

					processedDiffs = append(processedDiffs, equalStep, skipStep)
				} else {
					// here we need to break equality to two steps equal skip equal
					// TODO: implement the logic here
					processedDiffs = append(processedDiffs, processedDiff{
						Type:   DiffOperation(diffEqual),
						Items:  lines,
						Concat: prevConcat || concat,
					})
				}
			} else {
				processedDiffs = append(processedDiffs, processedDiff{
					Type:   DiffOperation(diffEqual),
					Items:  lines,
					Concat: prevConcat || concat,
				})
			}

		case diffInsert:
			processedDiffs = append(processedDiffs, processedDiff{
				Type:   DiffOperation(diffInsert),
				Items:  lines,
				Concat: prevConcat || concat,
			})
		case diffDelete:
			processedDiffs = append(processedDiffs, processedDiff{
				Type:   DiffOperation(diff.Type),
				Items:  lines,
				Concat: prevConcat || concat,
			})
		}
	}

	return processedDiffs
}

func padding(inserted, deleted int) (string, string) {
	digits := func(n int) (c int) {
		return len(strconv.Itoa(n))
	}

	i := digits(inserted)
	d := digits(deleted)
	if i == d {
		return "", ""
	}
	diff := i - d
	if diff > 0 {
		return "", strings.Repeat(" ", diff)
	}

	return strings.Repeat(" ", -diff), ""
}

func header(buff *bytes.Buffer, inserted, deleted int) {
	iPadding, dPadding := padding(inserted, deleted)

	buff.WriteString(newLine)
	buff.WriteString(diffDeleteText("- ", fmt.Sprintf("Snapshot %s - %d", dPadding, deleted), true))
	buff.WriteString(diffInsertText("+ ", fmt.Sprintf("Received %s + %d", iPadding, inserted), true))
	buff.WriteString(newLine)
}

func singleLineDiff(diffs []diffmatchpatch.Diff, inserted, deleted int) string {
	var a bytes.Buffer
	var b bytes.Buffer

	header(&a, inserted, deleted)

	a.WriteString(diffDeleteText("- ", "", false))
	b.WriteString(diffInsertText("+ ", "", false))

	for _, diff := range diffs {
		text, _, _ := strings.Cut(diff.Text, "\n")

		switch diff.Type {
		case diffDelete:
			a.WriteString(diffDeleteText("", coloredText(coloredText(text, white), boldRedBG), false))
		case diffInsert:
			b.WriteString(diffInsertText("", coloredText(coloredText(text, white), boldGreenBG), false))
		case diffEqual:
			a.WriteString(diffDeleteText("", text, false))
			b.WriteString(diffInsertText("", text, false))
		}
	}

	a.WriteString(newLine)
	b.WriteString(newLine)

	a.WriteString(b.String())

	return a.String()
}

func multiLineDiff(diffs []diffmatchpatch.Diff, inserted, deleted int) string {
	var buff bytes.Buffer
	pdiffs := processDiffLines(diffs)

	header(&buff, inserted, deleted)

	for _, diff := range pdiffs {
		lines := diff.Items
		newline := func(i int) bool {
			return (i+1 != len(lines) || !diff.Concat)
		}

		switch diff.Type {
		case EqualOp:
			for j, line := range lines {
				buff.WriteString(diffEqualText(stringTernary(diff.Concat, "", "  "), line, newline(j)))
			}
		case DeleteOp:
			for j, line := range lines {
				buff.WriteString(diffDeleteText(stringTernary(diff.Concat, "", "- "), line, newline(j)))
			}
		case InsertOp:
			for j, line := range lines {
				buff.WriteString(diffInsertText(stringTernary(diff.Concat, "", "+ "), line, newline(j)))
			}
		case SkipOp:
			// TODO: write this a bit better
			buff.WriteString(diffEqualText("", newLine+"---", true))
		}
	}

	return buff.String()
}

func isSingleLine(text string) bool {
	return strings.Count(text, newLine) == 1
}

func getStats(diffs []diffmatchpatch.Diff) (int, int) {
	var inserted, deleted int

	for _, d := range diffs {
		if d.Type == diffDelete {
			deleted++
		}

		if d.Type == diffInsert {
			inserted++
		}
	}
	return inserted, deleted
}

func prettyDiff(expected, received string) string {
	diffs := dmp.DiffCleanupSemanticLossless(
		dmp.DiffMain(expected, received, false),
	)
	if len(diffs) == 1 && diffs[0].Type == diffEqual {
		return ""
	}
	inserted, deleted := getStats(diffs)

	// NOTE: what happens if one of them is empty
	if isSingleLine(expected) && isSingleLine(received) {
		return singleLineDiff(diffs, inserted, deleted)
	}

	return multiLineDiff(diffs, inserted, deleted)
}
