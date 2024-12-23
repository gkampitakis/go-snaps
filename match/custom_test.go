package match

import (
	"errors"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestCustomMatcher(t *testing.T) {
	t.Run("should create a custom matcher", func(t *testing.T) {
		c := Custom("path", func(val any) (any, error) {
			return nil, nil
		})

		test.True(t, c.errOnMissingPath)
		test.Equal(t, c.path, "path")
		test.Equal(t, c.name, "Custom")
	})

	t.Run("should allow overriding values", func(t *testing.T) {
		c := Custom("path", func(val any) (any, error) {
			return nil, nil
		}).ErrOnMissingPath(false)

		test.False(t, c.errOnMissingPath)
		test.Equal(t, c.path, "path")
		test.Equal(t, c.name, "Custom")
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
			c := Custom("missing.key", func(val any) (any, error) {
				return nil, nil
			})

			res, errs := c.JSON(j)

			test.Nil(t, res)
			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "path does not exist", err.Reason.Error())
			test.Equal(t, "Custom", err.Matcher)
			test.Equal(t, "missing.key", err.Path)
		})

		t.Run("should ignore error in case of missing path", func(t *testing.T) {
			c := Custom("missing.key", func(val any) (any, error) {
				return nil, nil
			}).ErrOnMissingPath(false)

			res, errs := c.JSON(j)
			test.Equal(t, j, res)
			test.Nil(t, errs)
		})

		t.Run("should return error from custom callback", func(t *testing.T) {
			c := Custom("user.email", func(val any) (any, error) {
				return nil, errors.New("custom error")
			})

			res, errs := c.JSON(j)

			test.Nil(t, res)
			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "custom error", err.Reason.Error())
			test.Equal(t, "Custom", err.Matcher)
			test.Equal(t, "user.email", err.Path)
		})

		t.Run("should apply value from custom callback to json", func(t *testing.T) {
			c := Custom("user.email", func(val any) (any, error) {
				return "replaced email", nil
			})

			res, errs := c.JSON(j)

			expected := `{
			"user": {
				"name": "mock-user",
				"email": "replaced email"
			},
			"date": "16/10/2022"
		}`

			test.Equal(t, expected, string(res))
			test.Nil(t, errs)
		})
	})

	t.Run("YAML", func(t *testing.T) {
		y := []byte(`
user:
  name: mock-user
  email: mock-email
date: 16/10/2022
`)

		t.Run("should return error in case of missing path", func(t *testing.T) {
			c := Custom("$.missing.key", func(val any) (any, error) {
				return nil, nil
			})

			res, errs := c.YAML(y)

			test.Nil(t, res)
			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "path does not exist", err.Reason.Error())
			test.Equal(t, "Custom", err.Matcher)
			test.Equal(t, "$.missing.key", err.Path)
		})

		t.Run("should ignore error in case of missing path", func(t *testing.T) {
			c := Custom("$.missing.key", func(val any) (any, error) {
				return nil, nil
			}).ErrOnMissingPath(false)

			res, errs := c.YAML(y)
			test.Equal(t, y, res)
			test.Nil(t, errs)
		})

		t.Run("should return error from custom callback", func(t *testing.T) {
			c := Custom("$.user.email", func(val any) (any, error) {
				return nil, errors.New("custom error")
			})

			res, errs := c.YAML(y)

			test.Nil(t, res)
			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "custom error", err.Reason.Error())
			test.Equal(t, "Custom", err.Matcher)
			test.Equal(t, "$.user.email", err.Path)
		})

		t.Run("should apply value from custom callback to yaml", func(t *testing.T) {
			c := Custom("$.user.email", func(val any) (any, error) {
				return "replaced email", nil
			})

			res, errs := c.YAML(y)

			expected := `user:
  name: mock-user
  email: replaced email
date: 16/10/2022
`

			test.Equal(t, expected, string(res))
			test.Nil(t, errs)
		})
	})
}
