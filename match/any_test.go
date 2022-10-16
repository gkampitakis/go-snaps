package match

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestAnyMatcher(t *testing.T) {
	t.Run("should create an any matcher", func(t *testing.T) {
		p := []string{"test.1", "test.2"}
		a := Any(p...)

		test.True(t, a.errOnMissingPath)
		test.Equal(t, a.placeholder, "<Any value>")
		test.Equal(t, a.paths, p)
		test.Equal(t, a.name, "Any")
	})

	t.Run("should allow overriding values", func(t *testing.T) {
		p := []string{"test.1", "test.2"}
		a := Any(p...).ErrOnMissingPath(false).Placeholder("hello")

		test.False(t, a.errOnMissingPath)
		test.Equal(t, a.placeholder, "hello")
		test.Equal(t, a.paths, p)
		test.Equal(t, a.name, "Any")
	})

	t.Run("JSON", func(t *testing.T) {
		j := []byte(`{
			"user": {
				"name": "mock-user",
				"email": "mock-email"
			},
			"date": "16/10/2022"
		}`)

		t.Run("should return error in case of missing path", func(t *testing.T) {
			a := Any("user.2")
			res, errs := a.JSON(j)

			test.Equal(t, j, res)
			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "path does not exist", err.Reason.Error())
			test.Equal(t, "Any", err.Matcher)
			test.Equal(t, "user.2", err.Path)
		})

		t.Run("should aggregate errors", func(t *testing.T) {
			a := Any("user.2", "user.3")
			res, errs := a.JSON(j)

			test.Equal(t, j, res)
			test.Equal(t, 2, len(errs))
		})

		t.Run("should replace value and return new json", func(t *testing.T) {
			a := Any("user.email", "date", "missing.key").ErrOnMissingPath(false)
			res, errs := a.JSON(j)

			expected := `{
			"user": {
				"name": "mock-user",
				"email": "<Any value>"
			},
			"date": "<Any value>"
		}`

			test.Equal(t, 0, len(errs))
			test.Equal(t, expected, string(res))
		})

		t.Run("should replace value and return new json", func(t *testing.T) {
			a := Any(
				"user.email",
				"date",
				"missing.key",
			).ErrOnMissingPath(
				false,
			).Placeholder(10)
			res, errs := a.JSON(j)

			expected := `{
			"user": {
				"name": "mock-user",
				"email": 10
			},
			"date": 10
		}`

			test.Equal(t, 0, len(errs))
			test.Equal(t, expected, string(res))
		})
	})
}
