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

	var a, b string

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffEqual:
			a += text
			b += text
		case diffDelete:
			// str := stringTernary(spacesRegexp.MatchString(text), redBG(text), redText(text))
			// buff.WriteString(redBG(redText(text)))
			a += text
		case diffInsert:
			// str := stringTernary(spacesRegexp.MatchString(text), greenBG(text), greenText(text))
			// buff.WriteString(greenBG(greenText(text)))

			b += text
		}
	}

	buff.WriteString(redBG(redText(a)) + newLine + greenBG(greenText(b)))

	return buff.String()
}
