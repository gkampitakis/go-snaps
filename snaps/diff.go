package snaps

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	diffEqual  diffmatchpatch.Operation = 0
	diffInsert diffmatchpatch.Operation = 1
	diffDelete diffmatchpatch.Operation = -1
)

type processedDiff struct {
	Type   diffmatchpatch.Operation
	Items  []string
	Concat bool
}

// shows 11 lines 5 not changed 1 changed 5 not changed
// @@ -53,11 +53,11 @@
// https://stackoverflow.com/a/6508925/10068782
// createPatchMark

// TEST: how it looks with long snapshot more than 100 lines with small changes

func processDiffLines(diff []diffmatchpatch.Diff) []processedDiff {
	processedDiffs := []processedDiff{}

	for i, diff := range diff {
		text := diff.Text
		lines := strings.Split(text, newLine)
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
			processedDiffs = append(processedDiffs, processedDiff{
				Type:   diff.Type,
				Items:  lines,
				Concat: prevConcat || concat,
			})
		case diffInsert:
			processedDiffs = append(processedDiffs, processedDiff{
				Type:   diff.Type,
				Items:  lines,
				Concat: prevConcat || concat,
			})
		case diffDelete:
			processedDiffs = append(processedDiffs, processedDiff{
				Type:   diff.Type,
				Items:  lines,
				Concat: prevConcat || concat,
			})
		}
	}

	return processedDiffs
}

func header(buff *bytes.Buffer, inserted, deleted int) {
	buff.WriteString(newLine)
	buff.WriteString(diffDeleteText("- ", fmt.Sprintf("Snapshot  - %d ", deleted), true))
	buff.WriteString(diffInsertText("+ ", fmt.Sprintf("Received  + %d ", inserted), true))
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
		case diffEqual:
			for j, line := range lines {
				// TEST: with bug in missing new line
				// if line == "" {
				// 	line = "\n"
				// 	diff.Concat = false
				// }

				buff.WriteString(diffEqualText(stringTernary(diff.Concat, "", "  "), line, newline(j)))
			}
		case diffDelete:
			for j, line := range lines {
				// if line == "" {
				// 	line = "\n"
				// 	diff.Concat = false
				// }
				buff.WriteString(diffDeleteText(stringTernary(diff.Concat, "", "- "), line, newline(j)))
			}
		case diffInsert:
			for j, line := range lines {
				// if line == "" {
				// 	line = "\n"
				// 	diff.Concat = false
				// }
				buff.WriteString(diffInsertText(stringTernary(diff.Concat, "", "+ "), line, newline(j)))
			}
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
