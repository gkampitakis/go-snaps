package snaps

import (
	"bytes"

	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	diffEqual  diffmatchpatch.Operation = 0
	diffInsert diffmatchpatch.Operation = 1
	diffDelete diffmatchpatch.Operation = -1
)

func prettyDiff(expected, received string) string {
	diffs := dmp.DiffCleanupSemantic(dmp.DiffMain(expected, received, false))
	if len(diffs) == 1 && diffs[0].Type == diffEqual {
		return ""
	}

	var buff bytes.Buffer

	buff.WriteString(newLine)
	buff.WriteString(redBG(" Snapshot "))
	buff.WriteString(newLine)
	buff.WriteString(greenBG(" Received "))
	buff.WriteString(newLine + newLine)

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffEqual:
			buff.WriteString(dimText(text))
		case diffDelete:
			str := stringTernary(spacesRegexp.MatchString(text), redBG(text), redText(text))
			buff.WriteString(str)
		case diffInsert:
			str := stringTernary(spacesRegexp.MatchString(text), greenBG(text), greenText(text))
			buff.WriteString(str)
		}
	}

	buff.WriteString(newLine)

	return buff.String()
}
