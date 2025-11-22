package examples

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchInlineSnapshot(t *testing.T) {
	t.Run("should make an inline snapshot", func(t *testing.T) {
		u := struct {
			User string
			Age  int
		}{
			User: "mock-name",
			Age:  30,
		}

		snaps.MatchInlineSnapshot(
			t,
			u,
			snaps.Inline("struct { User string; Age int }{User:\"mock-name\", Age:30}"),
		)
	})

	t.Run("should create multiline inline snapshot", func(t *testing.T) {
		snaps.MatchInlineSnapshot(t, "line1\nline2\nline3", snaps.Inline(`line1
line2
line3`))
	})

	t.Run("should handle simple types", func(t *testing.T) {
		snaps.MatchInlineSnapshot(t, 42, snaps.Inline("int(42)"))
		snaps.MatchInlineSnapshot(t, 3.14159, snaps.Inline("float64(3.14159)"))
		snaps.MatchInlineSnapshot(t, true, snaps.Inline("bool(true)"))
		snaps.MatchInlineSnapshot(t, "hello", snaps.Inline("hello"))
	})

	t.Run("should handle slices and arrays", func(t *testing.T) {
		snaps.MatchInlineSnapshot(t, []int{1, 2, 3}, snaps.Inline("[]int{1, 2, 3}"))
		snaps.MatchInlineSnapshot(
			t,
			[]string{"a", "b", "c"},
			snaps.Inline("[]string{\"a\", \"b\", \"c\"}"),
		)
	})

	t.Run("should handle maps", func(t *testing.T) {
		m := map[string]int{"foo": 1, "bar": 2}
		snaps.MatchInlineSnapshot(t, m, snaps.Inline(`map[string]int{"bar":2, "foo":1}`))
	})

	t.Run("should handle nested structures", func(t *testing.T) {
		type Address struct {
			Street string
			City   string
		}
		type Person struct {
			Name    string
			Age     int
			Address Address
		}

		p := Person{
			Name: "John Doe",
			Age:  25,
			Address: Address{
				Street: "123 Main St",
				City:   "Springfield",
			},
		}

		snaps.MatchInlineSnapshot(t, p, snaps.Inline(`examples.Person{
    Name:    "John Doe",
    Age:     25,
    Address: examples.Address{Street:"123 Main St", City:"Springfield"},
}`))
	})

	t.Run("should handle pointers", func(t *testing.T) {
		val := 100
		ptr := &val
		snaps.MatchInlineSnapshot(t, ptr, snaps.Inline("&int(100)"))
	})

	t.Run("should handle empty values", func(t *testing.T) {
		snaps.MatchInlineSnapshot(t, "", snaps.Inline(""))
		snaps.MatchInlineSnapshot(t, []int{}, snaps.Inline("[]int{}"))
		snaps.MatchInlineSnapshot(t, map[string]int{}, snaps.Inline("map[string]int{}"))
	})

	t.Run("should handle special characters in strings", func(t *testing.T) {
		snaps.MatchInlineSnapshot(t, "hello\tworld", snaps.Inline("hello world"))
		snaps.MatchInlineSnapshot(t, "line1\nline2", snaps.Inline(`line1
line2`))
		snaps.MatchInlineSnapshot(t, "quotes: \"test\"", snaps.Inline("quotes: \"test\""))
		snaps.MatchInlineSnapshot(t, `quotes: "test"`, snaps.Inline("quotes: \"test\""))
	})

	t.Run("should handle multiple inline snapshots in sequence", func(t *testing.T) {
		type Product struct {
			ID    int
			Name  string
			Price float64
		}

		p1 := Product{ID: 1, Name: "Widget", Price: 9.99}
		p2 := Product{ID: 2, Name: "Gadget", Price: 19.99}

		snaps.MatchInlineSnapshot(
			t,
			p1,
			snaps.Inline("examples.Product{ID:1, Name:\"Widget\", Price:9.99}"),
		)
		snaps.MatchInlineSnapshot(
			t,
			p2,
			snaps.Inline("examples.Product{ID:2, Name:\"Gadget\", Price:19.99}"),
		)
	})

	t.Run("nested tests with inline snapshots", func(t *testing.T) {
		t.Run("inner test 1", func(t *testing.T) {
			snaps.MatchInlineSnapshot(t, "nested-1", snaps.Inline("nested-1"))
		})

		t.Run("inner test 2", func(t *testing.T) {
			snaps.MatchInlineSnapshot(t, "nested-2", snaps.Inline("nested-2"))
		})
	})

	t.Run("should handle complex multiline content", func(t *testing.T) {
		jsonLike := `{
  "name": "test",
  "value": 123,
  "nested": {
    "key": "value"
  }
}`
		snaps.MatchInlineSnapshot(t, jsonLike, snaps.Inline(`{
  "name": "test",
  "value": 123,
  "nested": {
    "key": "value"
  }
}`))
	})
}
