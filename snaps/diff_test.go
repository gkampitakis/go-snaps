package snaps

import (
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/internal/test"
)

var a = `Proin justo libero, pellentesque sit amet scelerisque ut, sollicitudin non tortor. 
		Sed velit elit, accumsan sed porttitor nec, elementum quis sapien. 
		Phasellus mattis purus in dui pretium, eu euismod metus feugiat. 
		Morbi turpis tellus, tincidunt mollis rutrum at, mattis laoreet lacus. 
		Donec in quam tempus, eleifend erat sit amet, aliquet metus. 
		Sed ullamcorper velit a est efficitur, et tempus ante rhoncus. 
		Aliquam diam sapien, vulputate sit amet elit sit amet, commodo eleifend sapien. 
		Donec consequat at nibh id mattis. Quisque vitae sagittis eros, convallis consectetur ante. 
		Duis finibus suscipit mi sed consectetur. Nulla libero neque, sagittis vel nulla vel,
		 vestibulum sagittis mauris. Ut laoreet urna lectus. 
		 Sed lorem felis, condimentum eget vehicula non, sagittis sit amet diam. 
		 Vivamus ut sapien at erat imperdiet suscipit id a lectus.`

var b = `Proin justo libero, pellentesque sit amet scelerisque ut, sollicitudin non tortor. 
		Sed velit elit, accumsan sed Ipsum nec, elementum quis sapien. 
		Phasellus mattis purus in dui pretium, eu euismod metus feugiat. 
		Morbi turpis Lorem, tincidunt mollis rutrum at, mattis laoreet lacus. 
		Donec in quam tempus, eleifend erat sit amet, aliquet metus. 
		Sed ullamcorper velit a est efficitur, et tempus ante rhoncus. 
		Aliquam diam sapien, vulputate sit amet elit sit amet, commodo eleifend sapien. 
		Donec consequat at nibh id mattis. Quisque vitae sagittis eros, convallis consectetur ante. 
		Duis finibus suscipit mi sed consectetur. Nulla libero neque, sagittis vel nulla vel,
		vestibulum sagittis mauris. Ut laoreet urna lectus. 
		Sed lorem felis, condimentum eget vehicula non, sagittis sit amet diam. 
		Vivamus ut sapien at erat imperdiet suscipit id a lectus.
		Another Line added.`

func TestStringUtils(t *testing.T) {
	t.Run("splitNewlines", func(t *testing.T) {
		for _, v := range []struct {
			input    string
			expected []string
		}{
			{"foo", []string{"foo\n"}},
			{"foo\nbar", []string{"foo\n", "bar\n"}},
			{"foo\nbar\n", []string{"foo\n", "bar\n", "\n"}},
			{`abc
			efg
			hello \n world`, []string{"abc\n", "\t\t\tefg\n", "\t\t\thello \\n world\n"}},
		} {
			v := v
			t.Run(v.input, func(t *testing.T) {
				t.Parallel()
				test.Equal(t, v.expected, splitNewlines(v.input))
			})
		}
	})

	t.Run("isSingleLine", func(t *testing.T) {
		test.Equal(t, true, isSingleline("hello world"))
		test.Equal(t, true, isSingleline("hello world\n"))
		test.Equal(t, false, isSingleline(`hello 
		 world
		 `))
		test.Equal(t, false, isSingleline("hello \n world\n"))
		test.Equal(t, false, isSingleline("hello \n world"))
	})
}

func TestDiff(t *testing.T) {
	t.Run("should return empty string if no diffs", func(t *testing.T) {
		t.Run("single line", func(t *testing.T) {
			expected, received := "Hello World\n", "Hello World\n"

			if diff := prettyDiff(expected, received); diff != "" {
				t.Errorf("found diff between same string %s", diff)
			}
		})

		t.Run("multiline", func(t *testing.T) {
			expected := `one snapshot
			containing new lines
			`
			received := expected

			if diff := prettyDiff(expected, received); diff != "" {
				t.Errorf("found diff between same string %s", diff)
			}
		})
	})

	t.Run("should print header consistently", func(t *testing.T) {
		MatchSnapshot(t, header(10000, 20))
		MatchSnapshot(t, header(20, 10000))
	})

	t.Run("with color", func(t *testing.T) {
		t.Run("should apply highlights on single line diff", func(t *testing.T) {
			a := strings.Repeat("abcd", 20)
			b := strings.Repeat("abcf", 20)

			MatchSnapshot(t, prettyDiff(a, b))
		})

		t.Run("multiline diff", func(t *testing.T) {
			MatchSnapshot(t, prettyDiff(a, b))
		})
	})

	t.Run("no color", func(t *testing.T) {
		t.Cleanup(func() {
			colors.NOCOLOR = false
		})
		colors.NOCOLOR = true

		t.Run("should apply highlights on single line diff", func(t *testing.T) {
			a := strings.Repeat("abcd", 20)
			b := strings.Repeat("abcf", 20)

			MatchSnapshot(t, prettyDiff(a, b))
		})

		t.Run("multiline diff", func(t *testing.T) {
			MatchSnapshot(t, prettyDiff(a, b))
		})
	})
}
