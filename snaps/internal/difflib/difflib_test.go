package difflib

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Equal(t *testing.T, expected, received interface{}) {
	t.Helper()

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("\n[expected]: %v\n[received]: %v", expected, received)
	}
}

func TestGetOptCodes(t *testing.T) {
	for _, v := range []struct {
		name     string
		a        string
		b        string
		expected []opCode
	}{
		{
			name: "qabxcd, abycdf",
			a:    "qabxcd",
			b:    "abycdf",
			expected: []opCode{
				{Tag: OpDelete, I1: 0, I2: 1, J1: 0, J2: 0},  // d a[0:1], (q) b[0:0] ()
				{Tag: OpEqual, I1: 1, I2: 3, J1: 0, J2: 2},   // e a[1:3], (ab) b[0:2] (ab)
				{Tag: OpReplace, I1: 3, I2: 4, J1: 2, J2: 3}, // r a[3:4], (x) b[2:3] (y)
				{Tag: OpEqual, I1: 4, I2: 6, J1: 3, J2: 5},   // e a[4:6], (cd) b[3:5] (cd)
				{Tag: OpInsert, I1: 6, I2: 6, J1: 5, J2: 6},  // i a[6:6], () b[5:6] (f)
			},
		},
		{
			name: "AsciiOnDelete",
			a:    strings.Repeat("a", 40) + "c" + strings.Repeat("b", 40),
			b:    strings.Repeat("a", 40) + strings.Repeat("b", 40),
			expected: []opCode{
				{OpEqual, 0, 40, 0, 40},
				{OpDelete, 40, 41, 40, 40},
				{OpEqual, 41, 81, 40, 80},
			},
		},
		{
			name: "AsciiOneInsert - 1",
			a:    strings.Repeat("b", 100),
			b:    "a" + strings.Repeat("b", 100),
			expected: []opCode{
				{OpInsert, 0, 0, 0, 1},
				{OpEqual, 0, 100, 1, 101},
			},
		},
		{
			name: "AsciiOneInsert - 2",
			a:    strings.Repeat("b", 100),
			b:    strings.Repeat("b", 50) + "a" + strings.Repeat("b", 50),
			expected: []opCode{
				{OpEqual, 0, 50, 0, 50},
				{OpInsert, 50, 50, 50, 51},
				{OpEqual, 50, 100, 51, 101},
			},
		},
	} {
		v := v
		t.Run(v.name, func(t *testing.T) {
			t.Parallel()

			a := strings.Split(v.a, "")
			b := strings.Split(v.b, "")
			Equal(t, v.expected, NewMatcher(a, b).getOpCodes())
		})
	}
}

func TestGroupedOpCodes(t *testing.T) {
	a := []string{}
	for i := 0; i != 39; i++ {
		a = append(a, fmt.Sprintf("%02d", i))
	}
	b := []string{}
	b = append(b, a[:8]...)
	b = append(b, " i")
	b = append(b, a[8:19]...)
	b = append(b, " x")
	b = append(b, a[20:22]...)
	b = append(b, a[27:34]...)
	b = append(b, " y")
	b = append(b, a[35:]...)
	s := NewMatcher(a, b)
	w := &strings.Builder{}
	for _, g := range s.GetGroupedOpCodes(-1) {
		fmt.Fprintf(w, "group\n")
		for _, op := range g {
			fmt.Fprintf(w, "  %d, %d, %d, %d, %d\n", op.Tag, op.I1, op.I2, op.J1, op.J2)
		}
	}
	expected := `group
  0, 5, 8, 5, 8
  1, 8, 8, 8, 9
  0, 8, 11, 9, 12
group
  0, 16, 19, 17, 20
  3, 19, 20, 20, 21
  0, 20, 22, 21, 23
  2, 22, 27, 23, 23
  0, 27, 30, 23, 26
group
  0, 31, 34, 27, 30
  3, 34, 35, 30, 31
  0, 35, 38, 31, 34
`

	Equal(t, expected, w.String())
}

func TestOutputFormatRangeFormatUnified(t *testing.T) {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	//
	// Each <range> field shall be of the form:
	//   %1d", <beginning line number>  if the range contains exactly one line,
	// and:
	//  "%1d,%1d", <beginning line number>, <number of lines> otherwise.
	// If a range is empty, its beginning line number shall be the number of
	// the line just before the range, or 0 if the empty range starts the file.
	Equal(t, "3,0", FormatRangeUnified(3, 3))
	Equal(t, "4", FormatRangeUnified(3, 4))
	Equal(t, "4,2", FormatRangeUnified(3, 5))
	Equal(t, "4,3", FormatRangeUnified(3, 6))
	Equal(t, "0,0", FormatRangeUnified(0, 0))
}
