package snaps

import (
	"bytes"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var dmp = diffmatchpatch.New()

func prettyPrintDiff(t *testing.T, expected, received string) string {
	diffs := dmp.DiffMain(expected, received, false)
	buff := bytes.Buffer{}

	if len(diffs) == 1 && diffs[0].Type == 0 {
		return ""
	}

	buff.WriteString("\n")
	buff.WriteString(redBG(" Expected "))
	buff.WriteString("\n")
	buff.WriteString(greenBG(" Received "))
	buff.WriteString("\n\n")

	for _, diff := range diffs {
		switch diff.Type {
		case 0:
			buff.WriteString(dimText(diff.Text))
		case -1:
			buff.WriteString(redBG(diff.Text))
		case 1:
			buff.WriteString(greenBG(diff.Text))
		}
	}

	buff.WriteString("\n")
	return buff.String()
}
