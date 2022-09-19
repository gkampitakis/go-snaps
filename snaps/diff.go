package snaps

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gkampitakis/go-diff/diffmatchpatch"
	"github.com/gkampitakis/go-snaps/snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/snaps/internal/difflib"
)

const (
	diffEqual  diffmatchpatch.Operation = 0
	diffInsert diffmatchpatch.Operation = 1
	diffDelete diffmatchpatch.Operation = -1
	context                             = 3
)

var dmp = diffmatchpatch.New()

func splitNewlines(s string) []string {
	lines := strings.SplitAfter(s, "\n")
	lines[len(lines)-1] += "\n"
	return lines
}

// isSingleline checks if a snapshot is one line or multiline.
// singleline snapshots have only one newline at the end.
func isSingleline(s string) bool {
	i := strings.Index(s, "\n")
	return i == len(s)-1 || i == -1
}

// shouldPrintHighlights checks if the two strings are going to be presented with
// inline highlights
func shouldPrintHighlights(a, b string) bool {
	return !colors.NOCOLOR && a != "\n" && b != "\n" && isSingleline(a) && isSingleline(b)
}

// Compare two sequences of lines; generate the delta as a unified diff.
//
// Unified diffs are a compact way of showing line changes and a few
// lines of context. The number of context lines is set by default to three.
//
// getUnifiedDiff returns a diff string along with inserted and deleted number.
func getUnifiedDiff(a, b string) (string, int, int) {
	aLines := splitNewlines(a)
	bLines := splitNewlines(b)

	var inserted int
	var deleted int
	var s strings.Builder

	s.Grow(len(a) + len(b))

	m := difflib.NewMatcher(aLines, bLines)
	for _, g := range m.GetGroupedOpCodes(context) {
		// aLines is a product of splitNewLines(), some items are just \"n"
		// if change is less than 10 items don't print the range
		if len(aLines) > 10 || len(bLines) > 10 {
			printRange(&s, g)
		}

		for _, c := range g {
			fallback := false
			i1, i2, j1, j2 := c.I1, c.I2, c.J1, c.J2

			if c.Tag == difflib.OpReplace {
				expected := strings.Join(bLines[j1:j2], "")
				received := strings.Join(aLines[i1:i2], "")

				if shouldPrintHighlights(expected, received) {
					i, d := singlelineDiff(&s, received, expected)
					inserted += i
					deleted += d

					continue
				}

				fallback = true
			}

			if c.Tag == difflib.OpEqual {
				for _, line := range aLines[i1:i2] {
					colors.FprintEqual(&s, line)
				}

				continue
			}

			// no continue, if fallback == true we want both lines printed
			if fallback || c.Tag == difflib.OpDelete {
				for _, line := range aLines[i1:i2] {
					colors.FprintDelete(&s, line)
					deleted++
				}
			}

			if fallback || c.Tag == difflib.OpInsert {
				for _, line := range bLines[j1:j2] {
					colors.FprintInsert(&s, line)
					inserted++
				}
			}
		}
	}

	return s.String(), inserted, deleted
}

// header of a diff report
//
// e.g.
//   - Snapshot - 10
//   - Received -  2
func header(inserted, deleted int) string {
	var s strings.Builder
	iPadding, dPadding := intPadding(inserted, deleted)

	s.WriteString("\n")
	colors.FprintDelete(&s, fmt.Sprintf("Snapshot %s- %d\n", dPadding, deleted))
	colors.FprintInsert(&s, fmt.Sprintf("Received %s+ %d\n", iPadding, inserted))
	s.WriteString("\n")

	return s.String()
}

func printRange(w io.Writer, opcodes []difflib.OpCode) {
	first, last := opcodes[0], opcodes[len(opcodes)-1]
	range1 := difflib.FormatRangeUnified(first.I1, last.I2)
	range2 := difflib.FormatRangeUnified(first.J1, last.J2)
	colors.FprintRange(w, range1, range2)
}

// IntPadding accepts two integers and returns two strings working as padding for aligning printed numbers
//
// e.g 1000 and 1 will return “ and "" ( 3 spaces ) so when printed will look
//
//	1000
//	   1
func intPadding(inserted, deleted int) (string, string) {
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

func singlelineDiff(s *strings.Builder, expected, received string) (int, int) {
	diffs := dmp.DiffCleanupSemantic(
		dmp.DiffMain(expected, received, false),
	)
	if len(diffs) == 1 && diffs[0].Type == diffEqual {
		// -1 means no diffs
		return -1, 0
	}

	var inserted, deleted int
	var i strings.Builder

	colors.FprintBg(s, colors.RedBg, colors.Reddiff, "- ")
	colors.FprintBg(&i, colors.GreenBG, colors.Greendiff, "+ ")

	for _, diff := range diffs {
		switch diff.Type {
		case diffDelete:
			deleted++
			colors.FprintDeleteBold(s, diff.Text)
		case diffInsert:
			inserted++
			colors.FprintInsertBold(&i, diff.Text)
		case diffEqual:
			colors.FprintBg(s, colors.RedBg, colors.Reddiff, diff.Text)
			colors.FprintBg(&i, colors.GreenBG, colors.Greendiff, diff.Text)
		}
	}

	s.WriteString(i.String())

	return inserted, deleted
}

func prettyDiff(expected, received string) string {
	if shouldPrintHighlights(expected, received) {
		var diff strings.Builder
		if i, d := singlelineDiff(&diff, expected, received); i != -1 {
			return header(i, d) + diff.String()
		}

		return ""
	}

	if diff, i, d := getUnifiedDiff(expected, received); diff != "" {
		return header(i, d) + diff
	}

	return ""
}
