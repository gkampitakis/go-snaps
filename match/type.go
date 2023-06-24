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

/*
Type matcher evaluates types that are passed in a snapshot

It replaces any targeted path with placeholder in the form of `<Type:ExpectedType>`

	match.Type[string]("user.info")
	// or with multiple paths
	match.Type[float64]("user.age", "data.items")
*/
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

		if _, ok := r.Value().(ExpectedType); !ok {
			errs = append(errs, MatcherError{
				Reason:  fmt.Errorf("expected type %T, received %T", *new(ExpectedType), r.Value()),
				Matcher: t.name,
				Path:    path,
			})

			continue
		}

		j, err := sjson.SetBytesOptions(
			json,
			path,
			fmt.Sprintf("<Type:%T>", r.Value()),
			&sjson.Options{
				Optimistic:     true,
				ReplaceInPlace: true,
			},
		)
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
