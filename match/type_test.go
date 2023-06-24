package match

import (
	"reflect"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestTypeMatcher(t *testing.T) {
	t.Run("should create a type matcher", func(t *testing.T) {
		p := []string{"test.1", "test.2"}
		tm := Type[string](p...)

		test.True(t, tm.errOnMissingPath)
		test.Equal(t, "Type", tm.name)
		test.Equal(t, p, tm.paths)
		test.Equal(t, reflect.TypeOf("").String(), reflect.TypeOf(tm.expectedType).String())
	})

	t.Run("should allow overriding values", func(t *testing.T) {
		p := []string{"test.1", "test.2"}
		tm := Type[string](p...)

		tm.ErrOnMissingPath(false)

		test.False(t, tm.errOnMissingPath)
		test.Equal(t, "Type", tm.name)
		test.Equal(t, p, tm.paths)
		test.Equal(t, reflect.TypeOf("").String(), reflect.TypeOf(tm.expectedType).String())
	})

	t.Run("JSON", func(t *testing.T) {
		j := []byte(`{
			"user": {
				"name": "mock-user",
				"email": "mock-email",
				"age": 29
			},
			"date": "16/10/2022"
		}`)

		t.Run("should return error in case of missing path", func(t *testing.T) {
			tm := Type[string]("user.2")
			res, errs := tm.JSON(j)

			test.Equal(t, j, res)
			test.Equal(t, 1, len(errs))

			err := errs[0]

			test.Equal(t, "path does not exist", err.Reason.Error())
			test.Equal(t, "Type", err.Matcher)
			test.Equal(t, "user.2", err.Path)
		})

		t.Run("should aggregate errors", func(t *testing.T) {
			tm := Type[string]("user.2", "user.3")
			res, errs := tm.JSON(j)

			test.Equal(t, j, res)
			test.Equal(t, 2, len(errs))
		})

		t.Run("should evaluate passed type and replace json", func(t *testing.T) {
			tm := Type[string]("user.name", "date")
			res, errs := tm.JSON(j)

			expected := `{
			"user": {
				"name": "<Type:string>",
				"email": "mock-email",
				"age": 29
			},
			"date": "<Type:string>"
		}`

			test.Nil(t, errs)
			test.Equal(t, expected, string(res))
		})

		t.Run("should return error with type mismatch", func(t *testing.T) {
			tm := Type[int]("user.name", "user.age")
			_, errs := tm.JSON(j)

			test.Equal(t, 2, len(errs))
			test.Equal(t, "expected type int, received string", errs[0].Reason.Error())
			test.Equal(t, "expected type int, received float64", errs[1].Reason.Error())
		})
	})
}
