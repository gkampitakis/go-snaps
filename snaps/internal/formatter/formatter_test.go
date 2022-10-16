package formatter

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"unsafe"
)

type test struct {
	v interface{}
	s string
}

type passtest struct {
	v    interface{}
	f, s string
}

type LongStructTypeName struct {
	longFieldName      interface{}
	otherLongFieldName interface{}
}

type SA struct {
	t *T
	v T
}

type T struct {
	x, y int
}

type F int

func (f F) Format(s fmt.State, c rune) {
	fmt.Fprintf(s, "F(%d)", int(f))
}

type Stringer struct{ i int }

func (s *Stringer) String() string { return "foo" }

var long = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var passthrough = []passtest{
	{1, "%d", "1"},
	{"a", "%s", "a"},
	{&Stringer{}, "%s", "foo"},
}

func TestPassthrough(t *testing.T) {
	for _, tt := range passthrough {
		s := fmt.Sprintf(tt.f, Formatter(tt.v))
		if tt.s != s {
			t.Errorf("expected %q", tt.s)
			t.Errorf("got      %q", s)
			t.Errorf("expraw\n%s", tt.s)
			t.Errorf("gotraw\n%s", s)
		}
	}
}

var gosyntax = []test{
	{nil, `nil`},
	{"", `""`},
	{"a", `"a"`},
	{1, "int(1)"},
	{1.0, "float64(1)"},
	{[]int(nil), "[]int(nil)"},
	{[0]int{}, "[0]int{}"},
	{complex(1, 0), "(1+0i)"},
	//{make(chan int), "(chan int)(0x1234)"},
	{unsafe.Pointer(uintptr(unsafe.Pointer(&long))), fmt.Sprintf("unsafe.Pointer(0x%02x)", uintptr(unsafe.Pointer(&long)))},
	{func(int) {}, "func(int) {...}"},
	{map[int]int{1: 1}, "map[int]int{1:1}"},
	{int32(1), "int32(1)"},
	{io.EOF, `&errors.errorString{s:"EOF"}`},
	{[]string{"a"}, `[]string{"a"}`},
	{
		[]string{long},
		`[]string{"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"}`,
	},
	{F(5), "formatter.F(5)"},
	{
		SA{&T{1, 2}, T{3, 4}},
		`formatter.SA{
    t:  &formatter.T{x:1, y:2},
    v:  formatter.T{x:3, y:4},
}`,
	},
	{
		map[int][]byte{1: {}},
		`map[int][]uint8{
    1:  {},
}`,
	},
	{
		map[int]T{1: {}},
		`map[int]formatter.T{
    1:  {},
}`,
	},
	{
		long,
		`"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"`,
	},
	{
		LongStructTypeName{
			longFieldName:      LongStructTypeName{},
			otherLongFieldName: long,
		},
		`formatter.LongStructTypeName{
    longFieldName:      formatter.LongStructTypeName{},
    otherLongFieldName: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
}`,
	},
	{
		&LongStructTypeName{
			longFieldName:      &LongStructTypeName{},
			otherLongFieldName: (*LongStructTypeName)(nil),
		},
		`&formatter.LongStructTypeName{
    longFieldName:      &formatter.LongStructTypeName{},
    otherLongFieldName: (*formatter.LongStructTypeName)(nil),
}`,
	},
	{
		[]LongStructTypeName{
			{nil, nil},
			{3, 3},
			{long, nil},
		},
		`[]formatter.LongStructTypeName{
    {},
    {
        longFieldName:      int(3),
        otherLongFieldName: int(3),
    },
    {
        longFieldName:      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
        otherLongFieldName: nil,
    },
}`,
	},
	{
		[]interface{}{
			LongStructTypeName{nil, nil},
			[]byte{1, 2, 3},
			T{3, 4},
			LongStructTypeName{long, nil},
		},
		`[]interface {}{
    formatter.LongStructTypeName{},
    []uint8{0x1, 0x2, 0x3},
    formatter.T{x:3, y:4},
    formatter.LongStructTypeName{
        longFieldName:      "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
        otherLongFieldName: nil,
    },
}`,
	},
}

func TestGoSyntax(t *testing.T) {
	for _, tt := range gosyntax {
		s := fmt.Sprintf("%# v", Formatter(tt.v))
		if tt.s != s {
			t.Errorf("expected %q", tt.s)
			t.Errorf("got      %q", s)
			t.Errorf("expraw\n%s", tt.s)
			t.Errorf("gotraw\n%s", s)
		}
	}
}

type I struct {
	i int
	R interface{}
}

func (i *I) I() *I { return i.R.(*I) }

func TestCycle(t *testing.T) {
	type A struct{ *A }
	v := &A{}
	v.A = v

	// panics from stack overflow without cycle detection
	t.Logf("Example cycle:\n%# v", Formatter(v))

	p := &A{}
	s := fmt.Sprintf("%# v", Formatter([]*A{p, p}))
	if strings.Contains(s, "CYCLIC") {
		t.Errorf("Repeated address detected as cyclic reference:\n%s", s)
	}

	type R struct {
		i int
		*R
	}
	r := &R{
		i: 1,
		R: &R{
			i: 2,
			R: &R{
				i: 3,
			},
		},
	}
	r.R.R.R = r
	t.Logf("Example longer cycle:\n%# v", Formatter(r))

	r = &R{
		i: 1,
		R: &R{
			i: 2,
			R: &R{
				i: 3,
				R: &R{
					i: 4,
					R: &R{
						i: 5,
						R: &R{
							i: 6,
							R: &R{
								i: 7,
								R: &R{
									i: 8,
									R: &R{
										i: 9,
										R: &R{
											i: 10,
											R: &R{
												i: 11,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	// here be pirates
	r.R.R.R.R.R.R.R.R.R.R.R = r
	t.Logf("Example very long cycle:\n%# v", Formatter(r))

	i := &I{
		i: 1,
		R: &I{
			i: 2,
			R: &I{
				i: 3,
				R: &I{
					i: 4,
					R: &I{
						i: 5,
						R: &I{
							i: 6,
							R: &I{
								i: 7,
								R: &I{
									i: 8,
									R: &I{
										i: 9,
										R: &I{
											i: 10,
											R: &I{
												i: 11,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	iv := i.I().I().I().I().I().I().I().I().I().I()
	*iv = *i
	t.Logf("Example long interface cycle:\n%# v", Formatter(i))
}

func TestStructFieldTags1(t *testing.T) {
	type testFieldTagCase struct {
		tc   interface{}
		want string
	}
	type testFieldTagT struct {
		z    string `go-snaps:"-"`
		x, y int
	}
	type testFieldTagT2 struct {
		x, y int `go-snaps:"-"`
		z    string
	}
	type testFieldTagT3 struct {
		x, y    int
		private string `go-snaps:"-"`
		a, b, c int
	}
	type testFieldTagT4 struct {
		fields []interface{}
	}
	for _, tc := range []testFieldTagCase{
		{tc: testFieldTagT{x: 1, y: 2, z: "ignored"}, want: `formatter.testFieldTagT{x:1, y:2}`},
		{tc: testFieldTagT2{x: 1, y: 2, z: "displayed"}, want: `formatter.testFieldTagT2{z:"displayed"}`},
		{tc: testFieldTagT3{x: 1, y: 2, private: "displayed", a: 42, b: 3, c: 4}, want: `formatter.testFieldTagT3{x:1, y:2, a:42, b:3, c:4}`},
		{tc: testFieldTagT4{fields: []interface{}{testFieldTagT3{x: 1, y: 2, private: "displayed", a: 42, b: 3, c: 4}}}, want: `formatter.testFieldTagT4{
    fields: {
        formatter.testFieldTagT3{x:1, y:2, a:42, b:3, c:4},
    },
}`},
	} {
		got := Sprint(tc.tc)
		if got != tc.want {
			t.Logf("Got:    %q", got)
			t.Logf("Wanted: %q", tc.want)
			t.Fail()
		}
	}
}
