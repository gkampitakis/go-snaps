package match

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestParsePath(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		path  string
		parts []string
	}{
		"simple path": {
			path:  "Hello",
			parts: []string{"Hello"},
		},
		"simple path signle letter": {
			path:  "H",
			parts: []string{"H"},
		},
		"simple path start .": {
			path:  ".Hello",
			parts: []string{"Hello"},
		},
		"simple path start . single letter": {
			path:  ".H",
			parts: []string{"H"},
		},
		"simple path end .": {
			path:  "Hello.",
			parts: []string{"Hello"},
		},
		"simple path start . and end .": {
			path:  ".Hello....",
			parts: []string{"Hello"},
		},
		"simple path start spaces": {
			path:  ".Hello.   ...",
			parts: []string{"Hello", "   "},
		},
		"multi nested path": {
			path:  "Hello.World.Path",
			parts: []string{"Hello", "World", "Path"},
		},
		"escape . inside key": {
			path:  `Hello.World\.Path`,
			parts: []string{"Hello", "World.Path"},
		},
		"escape . inside key at start": {
			path:  `\.Hello.World\.Path\.`,
			parts: []string{".Hello", "World.Path."},
		},
		"key with number": {
			path:  `1.\.Hello.World\.Path\..2`,
			parts: []string{"1", ".Hello", "World.Path.", "2"},
		},
		"key with slash but not escaping": {
			path:  ".hello.w\\orld",
			parts: []string{"hello", "w\\orld"},
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			test.Equal(t, tc.parts, parsePath(tc.path))
		})
	}
}
