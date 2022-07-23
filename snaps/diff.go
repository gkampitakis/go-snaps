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

// shows 11 lines 5 not changed 1 changed 5 not changed
// @@ -53,11 +53,11 @@
// https://stackoverflow.com/a/6508925/10068782
// createPatchMark

func getStats(diffs []diffmatchpatch.Diff) (added int, removed int) {
	for _, d := range diffs {
		if d.Type == diffInsert {
			added++
		}

		if d.Type == diffDelete {
			removed++
		}
	}

	return
}

// TODO: handle special case of one liner snapshot

func prettyDiff(expected, received string) string {
	// TODO: find color codes
	// TODO: update colors for the summary
	// TODO: ignore multiple diffs
	// TBD: Line summary ?

	// TEST: how it looks with long snapshot more than 100 lines with small changes
	diffs := dmp.DiffCleanupSemanticLossless(dmp.DiffMain(expected, received, false))
	if len(diffs) == 1 && diffs[0].Type == diffEqual {
		return ""
	}
	added, removed := getStats(diffs)

	var buff bytes.Buffer

	buff.WriteString(newLine)
	buff.WriteString(diffDeleteText("- ", fmt.Sprintf("Snapshot  - %d ", removed), true))
	buff.WriteString(diffInsertText("+ ", fmt.Sprintf("Received  + %d ", added), true))
	buff.WriteString(newLine)

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, newLine)
		newlinesuffix := lines[len(lines)-1] == ""
		if newlinesuffix {
			lines = lines[:len(lines)-1]
		}
		addNewLine := func(i int) bool {
			return (i+1 != len(lines) || newlinesuffix)
		}

		switch diff.Type {
		case diffEqual:
			for i, line := range lines {
				buff.WriteString(diffEqualText(stringTernary(len(lines) == 1, "", ""), line, addNewLine(i)))
			}
		case diffDelete:
			for i, line := range lines {

				buff.WriteString(diffDeleteText(stringTernary(len(lines) == 1, "", "- "), line, addNewLine(i)))
			}
		case diffInsert:
			for i, line := range lines {
				buff.WriteString(diffInsertText(stringTernary(len(lines) == 1, "", "+ "), line, addNewLine(i)))
			}
		}
	}

	return buff.String()
}
