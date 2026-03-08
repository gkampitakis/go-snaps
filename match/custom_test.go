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

	t.Run("should allow overriding config values", func(t *testing.T) {
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

			_, errs := c.JSON(j)

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

			_, errs := c.JSON(j)

			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "custom error", err.Reason.Error())
			test.Equal(t, "Custom", err.Matcher)
			test.Equal(t, "user.email", err.Path)
		})

		t.Run("should apply value from custom callback to json", func(t *testing.T) {
			c := Custom("user.email", func(val any) (any, error) {
				test.Equal(t, "mock-email", val.(string))
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

		t.Run("nested json arrays", func(t *testing.T) {
			t.Run("should replace values with root level nested arrays", func(t *testing.T) {
				j := []byte(`[
					{
						"results": ["mock-data-1", "mock-data-2" ],
					},
					{
						"results": ["mock-data-1", "mock-data-2" ],
					},
					{
						"results": ["mock-data-1", "mock-data-2" ],
					},
				]`)

				a := Custom("#.results.#", func(val any) (any, error) {
					return "<custom>", nil
				})

				res, errs := a.JSON(j)

				expected := `[
					{
						"results": ["<custom>", "<custom>" ],
					},
					{
						"results": ["<custom>", "<custom>" ],
					},
					{
						"results": ["<custom>", "<custom>" ],
					},
				]`

				test.Equal(t, 0, len(errs))
				test.Equal(t, expected, string(res))
			})

			t.Run("should replace value and return new json", func(t *testing.T) {
				j := []byte(`{
					"results": [
						{
							"packages": [
								{"vulnerabilities": "mock-data-1", "name": "mock-name-1", "id": 12},
								{"vulnerabilities": "mock-data-1", "name": "mock-name-1", "id": 15},
								{"vulnerabilities": "mock-data-1", "name": "mock-name-1", "id": 17},
							],
						},
						{
							"packages": [
								{"vulnerabilities": "mock-data-2", "name": "mock-name-2", "id": 22},
								{"vulnerabilities": "mock-data-2", "name": "mock-name-2", "id": 25},
								{"vulnerabilities": "mock-data-2", "name": "mock-name-2", "id": 27},
							],
						},
						{
							"packages": [
								{"vulnerabilities": "mock-data-3", "name": "mock-name-3", "id": 32},
								{"vulnerabilities": "mock-data-3", "name": "mock-name-3", "id": 35},
								{"vulnerabilities": "mock-data-3", "name": "mock-name-3", "id": 37},
							],
						},
					]
				}`)
				a := Custom(
					"results.#.packages.#.vulnerabilities",
					func(val any) (any, error) {
						return "<custom>", nil
					},
				).ErrOnMissingPath(false)
				res, errs := a.JSON(j)

				expected := `{
					"results": [
						{
							"packages": [
								{"vulnerabilities": "<custom>", "name": "mock-name-1", "id": 12},
								{"vulnerabilities": "<custom>", "name": "mock-name-1", "id": 15},
								{"vulnerabilities": "<custom>", "name": "mock-name-1", "id": 17},
							],
						},
						{
							"packages": [
								{"vulnerabilities": "<custom>", "name": "mock-name-2", "id": 22},
								{"vulnerabilities": "<custom>", "name": "mock-name-2", "id": 25},
								{"vulnerabilities": "<custom>", "name": "mock-name-2", "id": 27},
							],
						},
						{
							"packages": [
								{"vulnerabilities": "<custom>", "name": "mock-name-3", "id": 32},
								{"vulnerabilities": "<custom>", "name": "mock-name-3", "id": 35},
								{"vulnerabilities": "<custom>", "name": "mock-name-3", "id": 37},
							],
						},
					]
				}`

				test.Equal(t, 0, len(errs))
				test.Equal(t, expected, string(res))
			})
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
