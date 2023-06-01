package match

import (
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type typeMatcher[ExpectedType any] struct {
	paths            []string
	errOnMissingPath bool
	name             string
	expectedType     interface{}
}

func Type[ExpectedType any](paths ...string) *typeMatcher[ExpectedType] {
	return &typeMatcher[ExpectedType]{
		paths:            paths,
		errOnMissingPath: true,
		name:             "Type",
		expectedType:     *new(ExpectedType),
	}
}

// ErrOnMissingPath determines if matcher will fail in case of trying to access a json path
// that doesn't exist
func (t *typeMatcher[T]) ErrOnMissingPath(e bool) *typeMatcher[T] {
	t.errOnMissingPath = e
	return t
}

func (t typeMatcher[ExpectedType]) JSON(s []byte) ([]byte, []MatcherError) {
	var errs []MatcherError
	json := s

	for _, path := range t.paths {
		r := gjson.GetBytes(json, path)
		if !r.Exists() {
			if t.errOnMissingPath {
				errs = append(errs, MatcherError{
					Reason:  errors.New("path does not exist"),
					Matcher: t.name,
					Path:    path,
				})
			}
			continue
		}

		value := fmt.Sprintf("<Type:%T>", *new(ExpectedType))
		if _, ok := r.Value().(ExpectedType); !ok {
			value = fmt.Sprintf("<Type:%T>", r.Value())
		}

		j, err := sjson.SetBytesOptions(json, path, value, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})
		if err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: t.name,
				Path:    path,
			})

			continue
		}

		json = j
	}

	return json, errs
}
