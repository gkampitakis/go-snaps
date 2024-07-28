package examples

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchStandaloneSnapshot(t *testing.T) {
	t.Run("should create html snapshots", func(t *testing.T) {
		snaps := snaps.WithConfig(
			snaps.Ext(".html"),
		)

		snaps.MatchStandaloneSnapshot(t, `
<!DOCTYPE html>
<html>
<body>

<h1>My First Heading</h1>

<p>My first paragraph.</p>

</body>
</html>
`)
		snaps.MatchStandaloneSnapshot(t, "<div>Hello World</div>")
	})

	t.Run("should create standalone snapshots with specified filename", func(t *testing.T) {
		snaps := snaps.WithConfig(
			snaps.Filename("my_standalone_snap"),
		)

		snaps.MatchStandaloneSnapshot(t, "hello world-0")
		snaps.MatchStandaloneSnapshot(t, "hello world-1")
		snaps.MatchStandaloneSnapshot(t, "hello world-2")
		snaps.MatchStandaloneSnapshot(t, "hello world-3")
	})
}
