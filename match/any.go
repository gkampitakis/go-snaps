package match

import (
	"errors"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths            []string
	placeholder      string
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
		paths:            paths,
		placeholder:      "<Any value>",
		errOnMissingPath: false,
		name:             "Any",
	}
}

// Placeholder allows to define the placeholder string for Any matcher
func (a *anyMatcher) Placeholder(p string) *anyMatcher {
	a.placeholder = p
	return a
}

// ErrOnMissingPath will make the fail in case a path accessed doesn't exist
func (a *anyMatcher) ErrOnMissingPath() *anyMatcher {
	a.errOnMissingPath = true
	return a
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Any matchers
func (a anyMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	var merrors []MatcherError

	newJSON := s
	for _, path := range a.paths {
		r := gjson.GetBytes(newJSON, path)
		if !r.Exists() {
			if a.errOnMissingPath {
				merrors = append(merrors, MatcherError{
					Reason:  errors.New("path does not exist"),
					Matcher: a.name,
					Path:    path,
				})
			}
			continue
		}

		j, err := sjson.SetBytesOptions(newJSON, path, a.placeholder, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})
		if err != nil {
			merrors = append(merrors, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    path,
			})

			continue
		}

		newJSON = j
	}

	return newJSON, merrors
}
