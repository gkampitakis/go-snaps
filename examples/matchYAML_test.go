package examples

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/match"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMatchYaml(t *testing.T) {
	t.Run("should match string yaml", func(t *testing.T) {
		const doc_yaml = `
ser: "mock-service"
name: John Doe
age: 30
email: john.doe@example.com
address: "123 Main St"
street: 123 Main St`

		snaps.WithConfig(snaps.Update(os.Getenv("UPDATE_ONLY") == "true")).
			// .address panics
			MatchYAML(t, doc_yaml, match.Any("$.address", "$.street").Placeholder("mock-address"))
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

		snaps.WithConfig(snaps.Update(os.Getenv("UPDATE_ONLY") == "true")).
			MatchYAML(t, User{
				Service: "mock-service",
				Name:    "John Doe",
				Age:     30,
				Email:   "john.doe@example.com",
				Address: "123 Main St",
				Street:  "123 Main St",
			}, match.Any("$.address", "$.street").Placeholder("mock-address"))
	})
}
