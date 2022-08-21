package snaps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gkampitakis/go-snaps/snaps/internal"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	diffEqual  diffmatchpatch.Operation = 0
	diffInsert diffmatchpatch.Operation = 1
	diffDelete diffmatchpatch.Operation = -1
	context                             = 3
)

var (
	dmp = diffmatchpatch.New()
)

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

	m := internal.NewMatcher(aLines, bLines)
	for _, g := range m.GetGroupedOpCodes(context) {
		// aLines is a product of splitNewLines(), some items are just \"n"
		// if change is less than 10 items don't print the range
		if len(aLines) > 10 || len(bLines) > 10 {
			first, last := g[0], g[len(g)-1]
			range1 := internal.FormatRangeUnified(first.I1, last.I2)
			range2 := internal.FormatRangeUnified(first.J1, last.J2)
			fprintRange(&s, range1, range2)
		}

		for _, c := range g {
			fallback := false
			i1, i2, j1, j2 := c.I1, c.I2, c.J1, c.J2

			if c.Tag == internal.OpReplace {
				expected := strings.Join(bLines[j1:j2], "")
				received := strings.Join(aLines[i1:i2], "")

				if expected != "\n" && received != "\n" &&
					isSingleline(received) && isSingleline(expected) {
					i, d := singlelineDiff(&s, received, expected)
					inserted += i
					deleted += d

					continue
				}

				fallback = true
			}

			if c.Tag == internal.OpEqual {
				for _, line := range aLines[i1:i2] {
					fprintEqual(&s, line)
				}

				continue
			}

			// no continue, if fallback = true we want both lines printed
			if fallback || c.Tag == internal.OpDelete {
				for _, line := range aLines[i1:i2] {
					fprintDelete(&s, line)
					deleted++
				}
			}

			if fallback || c.Tag == internal.OpInsert {
				for _, line := range bLines[j1:j2] {
					fprintInsert(&s, line)
					inserted++
				}
			}
		}
	}

	if s.Len() == 0 {
		// -1 means no diffs
		return "", -1, 0
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
	fprintDelete(&s, fmt.Sprintf("Snapshot %s- %d\n", dPadding, deleted))
	fprintInsert(&s, fmt.Sprintf("Received %s+ %d\n", iPadding, inserted))
	s.WriteString("\n")

	return s.String()
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
		s.Reset()
		// -1 means no diffs
		return -1, 0
	}

	var inserted, deleted int
	var i strings.Builder

	fprintBgColored(s, redBG, reddiff, "- ")
	fprintBgColored(&i, greenBG, greendiff, "+ ")

	for _, diff := range diffs {
		switch diff.Type {
		case diffDelete:
			deleted++
			fprintDeleteBold(s, diff.Text)
		case diffInsert:
			inserted++
			fprintInsertBold(&i, diff.Text)
		case diffEqual:
			fprintBgColored(s, redBG, reddiff, diff.Text)
			fprintBgColored(&i, greenBG, greendiff, diff.Text)
		}
	}

	s.WriteString(i.String())

	return inserted, deleted
}

func prettyDiff(expected, received string) string {
	if isSingleline(expected) && isSingleline(received) {
		var diff strings.Builder
		if i, d := singlelineDiff(&diff, expected, received); i != -1 {
			return header(i, d) + diff.String()
		}

		return ""
	}

	if diff, i, d := getUnifiedDiff(expected, received); i != -1 {
		return header(i, d) + diff
	}

	return ""
}
