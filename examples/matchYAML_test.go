package examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchYaml(t *testing.T) {
	t.Run("should match struct yaml", func(t *testing.T) {
		type User struct {
			Name    string    `yaml:"name"`
			Age     int       `yaml:"age"`
			Email   string    `yaml:"email"`
			Address string    `yaml:"address"`
			Time    time.Time `yaml:"time"`
		}

		snaps.MatchYAML(t, User{
			Name:    "John Doe",
			Age:     30,
			Email:   "john.doe@example.com",
			Address: "123 Main St",
			Time:    time.Now(),
		}, match.Any("$.time").Placeholder("mock-time"), match.Any("$.address").Placeholder("mock-address"))
	})

	t.Run("custom matching logic", func(t *testing.T) {
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

		snaps.MatchYAML(t, u, match.Custom("$.keys", func(val any) (any, error) {
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

	t.Run("type matcher", func(t *testing.T) {
		snaps.MatchYAML(t, "data: 10", match.Type[uint64]("$.data"))

		snaps.MatchYAML(
			t,
			"metadata:\n timestamp: 1687108093142",
			match.Type[map[string]any]("$.metadata"),
		)
	})
}
