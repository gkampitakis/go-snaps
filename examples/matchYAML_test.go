package examples

import (
	"os"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

var my_snaps = snaps.WithConfig(snaps.Update(os.Getenv("UPDATE_ONLY") == "true"))

func TestMatchYaml(t *testing.T) {
	t.Run("should match string yaml", func(t *testing.T) {
		const doc_yaml = `
ser: "mock-service"
name: John Doe
age: 30
email: john.doe@example.com
address: "123 Main St"
street: 123 Main St`

		my_snaps.MatchYAML(
			t,
			doc_yaml,
			match.Any("$.address", "$.street").Placeholder("mock-address"),
		)
	})

	t.Run("should match struct yaml", func(t *testing.T) {
		type User struct {
			Service string `yaml:"ser"`
			Name    string `yaml:"name"`
			Age     int    `yaml:"age"`
			Email   string `yaml:"email"`
			Address string `yaml:"address"`
			Street  string `yaml:"street"`
		}

		my_snaps.MatchYAML(t, User{
			Service: "mock-service",
			Name:    "John Doe",
			Age:     30,
			Email:   "john.doe@example.com",
			Address: "123 Main St",
			Street:  "123 Main St",
		}, match.Any("$.address", "$.street").Placeholder("mock-address"))
	})

	t.Run("should fail with invalid yaml input", func(t *testing.T) {
		snaps.SkipNow(t)

		doc_yaml := []byte(`
ser: "mock-service"
name: John Doe
age: 30
email: john.doe@example.com
address:
 - 1
 value: test
street: 123 Main St`)

		my_snaps.MatchYAML(t, doc_yaml)
	})

	t.Run("should assert type", func(t *testing.T) {
		doc_yaml := []byte(`
ser: "mock-service"
name: John Doe
age: 30
f: 3.5
b: !!bool true
email: john.doe@example.com
street: 123 Main St`)

		my_snaps.MatchYAML(t, doc_yaml,
			match.Type[uint64]("$.age"),
			match.Type[string]("$.email"),
			match.Type[bool]("$.b"),
			match.Type[float64]("$.f"),
		)
	})

	t.Run("custom matcher", func(t *testing.T) {
		doc_yaml := []byte(`
ser: "mock-service"
name: John Doe
age: 30
f: 3.5
b: !!bool true
email: john.doe@example.com
street: 123 Main St`)

		my_snaps.MatchYAML(t, doc_yaml, match.Custom("$.email", func(val any) (any, error) {
			return strings.ToUpper(val.(string)), nil
		}))
	})
}
