package examples

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchJSON(t *testing.T) {
	t.Run("should make a json object snapshot", func(t *testing.T) {
		m := map[string]interface{}{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": struct{ Msg string }{"Hello World"},
			"mock-3": float32(10.4),
		}

		snaps.MatchJSON(t, m)
	})

	t.Run("should create a prettyJSON snap", func(t *testing.T) {
		value := `{"user":"mock-user","age":10,"email":"mock@email.com"}`
		snaps.MatchJSON(t, value)
		snaps.MatchJSON(t, []byte(value))
	})

	t.Run("should marshal struct", func(t *testing.T) {
		type User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Keys  []int  `json:"keys"`
		}

		u := User{
			Name:  "mock-name",
			Email: "mock@email.com",
			Keys:  []int{1, 2, 3, 4, 5},
		}

		snaps.MatchJSON(t, u)
	})
}
