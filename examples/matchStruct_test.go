package examples_test

import (
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

type User struct {
	Name     string
	Age      int
	internal string
	Data     map[string]int
	Nested   struct {
		Metadata string
		Date     time.Time
	}
}

func TestMatchStruct(t *testing.T) {
	user := User{Name: "george", Age: 10, internal: "metadatav2",
		Nested: struct {
			Metadata string
			Date     time.Time
		}{
			Metadata: "some more data here",
			Date:     time.Now(),
		}}

	t.Run("should create a simple snapshot", func(t *testing.T) {
		//  "internal"
		snaps.MatchStruct(t, &user, match.Any("Nested").ErrOnMissingPath(false))
	})
}
