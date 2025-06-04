package examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchStandaloneYAML(t *testing.T) {
	t.Run("should make a yaml object snapshot", func(t *testing.T) {
		m := map[string]any{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": struct{ Msg string }{"Hello World"},
			"mock-3": float32(10.4),
		}

		snaps.MatchStandaloneYAML(t, m)
	})

	t.Run("should marshal struct", func(t *testing.T) {
		type User struct {
			Name  string   `yaml:"name"`
			Email string   `yaml:"email"`
			Keys  []int    `yaml:"keys"`
			Tags  []string `yaml:"tags"`
		}

		u := User{
			Name:  "mock-name",
			Email: "mock@email.com",
			Keys:  []int{1, 2, 3, 4, 5},
		}

		snaps.MatchStandaloneYAML(t, u)
	})

	t.Run("matchers", func(t *testing.T) {
		t.Run("Custom matcher", func(t *testing.T) {
			t.Run("struct marshalling", func(t *testing.T) {
				type User struct {
					Name  string `yaml:"name"`
					Email string `yaml:"email"`
					Keys  []int  `yaml:"keys"`
				}

				u := User{
					Name:  "mock-name",
					Email: "mock-user@email.com",
					Keys:  []int{1, 2, 3, 4, 5},
				}

				snaps.MatchStandaloneYAML(t, u, match.Custom("$.keys", func(val any) (any, error) {
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

			t.Run("YAML string validation", func(t *testing.T) {
				value := `user: mock-user
age: 2
email: mock@email.com`

				snaps.MatchStandaloneYAML(
					t,
					value,
					match.Custom("$.age", func(val any) (any, error) {
						if valInt, ok := val.(uint64); !ok || valInt >= 5 {
							return nil, fmt.Errorf("expecting number less than 5")
						}

						return "<less than 5 age>", nil
					}),
				)
			})
		})

		t.Run("Any matcher", func(t *testing.T) {
			t.Run("should ignore fields", func(t *testing.T) {
				value := fmt.Sprintf(`user: mock-user
age: 10
nested:
 now:
 - "%s"`, time.Now())
				snaps.MatchStandaloneYAML(t, value, match.Any("$.nested.now[0]"))
			})
		})

		t.Run("Type matcher", func(t *testing.T) {
			t.Run("should create snapshot with type placeholder", func(t *testing.T) {
				snaps.MatchStandaloneYAML(t, `data: 10`, match.Type[uint64]("$.data"))
				snaps.MatchStandaloneYAML(
					t,
					`metadata:
 timestamp: 1687108093142
`,
					match.Type[map[string]any]("$.metadata"),
				)
			})
		})
	})
}
