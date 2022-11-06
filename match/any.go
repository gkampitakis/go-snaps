package match

import (
	"errors"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths            []string
	placeholder      interface{}
	errOnMissingPath bool
	name             string
}

/*
Any matcher acts as a placeholder for any value

It replaces any targeted path with a placeholder string

	Any("user.name")
	// or with multiple paths
	Any("user.name", "user.email")
*/
func Any(paths ...string) *anyMatcher {
	return &anyMatcher{
		errOnMissingPath: true,
		placeholder:      "<Any value>",
		paths:            paths,
		name:             "Any",
	}
}

// Placeholder allows to define the placeholder value for Any matcher
func (a *anyMatcher) Placeholder(p interface{}) *anyMatcher {
	a.placeholder = p
	return a
}

// ErrOnMissingPath determines if matcher will fail in case of trying to access a json path
// that doesn't exist
func (a *anyMatcher) ErrOnMissingPath(e bool) *anyMatcher {
	a.errOnMissingPath = e
	return a
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Any matchers
func (a anyMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	json := s
	for _, path := range a.paths {
		r := gjson.GetBytes(json, path)
		if !r.Exists() {
			if a.errOnMissingPath {
				errs = append(errs, MatcherError{
					Reason:  errors.New("path does not exist"),
					Matcher: a.name,
					Path:    path,
				})
			}
			continue
		}

		j, err := sjson.SetBytesOptions(json, path, a.placeholder, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})
		if err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    path,
			})

			continue
		}

		json = j
	}

	return json, errs
}
