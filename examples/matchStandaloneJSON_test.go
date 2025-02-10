package examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchStandaloneJSON(t *testing.T) {
	t.Run("should make a json object snapshot", func(t *testing.T) {
		m := map[string]any{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": struct{ Msg string }{"Hello World"},
			"mock-3": float32(10.4),
		}

		snaps.MatchStandaloneJSON(t, m)
	})

	t.Run("should create a prettyJSON snap", func(t *testing.T) {
		value := `{"user":"mock-user","age":10,"email":"mock@email.com"}`
		snaps.MatchStandaloneJSON(t, value)
		snaps.MatchStandaloneJSON(t, []byte(value))
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

		snaps.MatchStandaloneJSON(t, u)
	})

	t.Run("should create a prettyJSON snap following by config", func(t *testing.T) {
		value := `{"user":"mock-user","age":10,"email":"mock@email.com"}`
		config := snaps.WithConfig(snaps.JSON(snaps.JSONConfig{
			Indent:   "    ",
			SortKeys: false,
		}))
		config.MatchStandaloneJSON(t, value)
		config.MatchStandaloneJSON(t, []byte(value))
	})

	t.Run("matchers", func(t *testing.T) {
		t.Run("Custom matcher", func(t *testing.T) {
			t.Run("struct marshalling", func(t *testing.T) {
				type User struct {
					Name  string `json:"name"`
					Email string `json:"email"`
					Keys  []int  `json:"keys"`
				}

				u := User{
					Name:  "mock-user",
					Email: "mock-user@email.com",
					Keys:  []int{1, 2, 3, 4, 5},
				}

				snaps.MatchStandaloneJSON(t, u, match.Custom("keys", func(val any) (any, error) {
					keys, ok := val.([]any)
					if !ok {
						return nil, fmt.Errorf("expected []any but got %T", val)
					}

					if len(keys) > 5 {
						return nil, fmt.Errorf("expected less than 5 keys")
					}

					return val, nil
				}))
			})

			t.Run("JSON string validation", func(t *testing.T) {
				value := `{"user":"mock-user","age":2,"email":"mock@email.com"}`

				snaps.MatchStandaloneJSON(t, value, match.Custom("age", func(val any) (any, error) {
					if valInt, ok := val.(float64); !ok || valInt >= 5 {
						return nil, fmt.Errorf("expecting number less than 5")
					}

					return "<less than 5 age>", nil
				}))
			})
		})

		t.Run("Any matcher", func(t *testing.T) {
			t.Run("should ignore fields", func(t *testing.T) {
				value := fmt.Sprintf(
					`{"user":"mock-user","age":10,"nested":{"now":["%s"]}}`,
					time.Now(),
				)
				snaps.MatchStandaloneJSON(t, value, match.Any("nested.now.0"))
			})
		})

		t.Run("Type matcher", func(t *testing.T) {
			t.Run("should create snapshot with type placeholder", func(t *testing.T) {
				snaps.MatchStandaloneJSON(t, `{"data":10}`, match.Type[float64]("data"))
				snaps.MatchStandaloneJSON(
					t,
					`{"metadata":{"timestamp":"1687108093142"}}`,
					match.Type[map[string]any]("metadata"),
				)
			})
		})
	})
}
