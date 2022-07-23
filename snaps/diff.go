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

func prettyDiff(expected, received string) string {

	res := strings.Split(expected, "\n")
	res2 := strings.Split(received, "\n")

	fmt.Println(res, res2, strings.Title(expected))

	diffs := dmp.DiffCleanupSemantic(dmp.DiffMain(res[0], res2[0], false))
	diffs := dmp.DiffCleanupSemanticLossless(dmp.DiffMain(expected, received, false))
	if len(diffs) == 1 && diffs[0].Type == diffEqual {
		return ""
	}

	added, removed := getStats(diffs)

	var buff bytes.Buffer

	buff.WriteString(newLine)
	// FIXME:
	buff.WriteString(redBG(redText(fmt.Sprintf("- Snapshot - %d ", removed))))
	buff.WriteString(newLine)
	buff.WriteString(greenBG(greenText(fmt.Sprintf("+ Received + %d ", added))))
	buff.WriteString(newLine + newLine)

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		addNewLine := lines[len(lines)-1] == ""
		if addNewLine {
			lines = lines[:len(lines)-1]
		}

		switch diff.Type {
		case diffEqual:
			for i, line := range lines {
				if i+1 == len(lines) && !addNewLine {
					buff.WriteString(" " + dimText(line))
					continue
				}

				buff.WriteString(" " + dimText(line) + newLine)
			}
		case diffDelete:
			for i, line := range lines {
				if i+1 == len(lines) && !addNewLine {
					buff.WriteString(redBG(redText(fmt.Sprintf("- %s", line))))
					continue
				}

				buff.WriteString(redBG(redText(fmt.Sprintf("- %s", line))) + newLine)
			}

		case diffInsert:
			for i, line := range lines {
				if i+1 == len(lines) && !addNewLine {
					buff.WriteString(greenBG(greenText(fmt.Sprintf("+ %s", line))))
					continue
				}

				buff.WriteString(greenBG(greenText(fmt.Sprintf("+ %s", line))) + newLine)

			}
		}
	}

	return buff.String()
}
