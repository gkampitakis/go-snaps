package match

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths            []string
	placeholder      string
	errOnMissingPath bool
	name             string
}

// NOTE: order matters
// TODO: add comments on public fns
func Any(paths ...string) *anyMatcher {
	return &anyMatcher{
		paths:            paths,
		placeholder:      "<Any value>",
		errOnMissingPath: false,
		name:             "Any",
	}
}

func (a *anyMatcher) Placeholder(p string) *anyMatcher {
	a.placeholder = p
	return a
}

func (a *anyMatcher) ErrOnMissingPath() *anyMatcher {
	a.errOnMissingPath = true
	return a
}

// NOTE: we need to finalize the JSONMatcher return values
// Finalize the any functionality
// the matcher needs to be extensible. If not it's done

// internal method
func (a anyMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	var errors []MatcherError

	newJSON := s
	for _, path := range a.paths {
		r := gjson.GetBytes(newJSON, path)
		if !r.Exists() {
			if a.errOnMissingPath {
				errors = append(errors, MatcherError{
					Reason:  "path does not exist",
					Matcher: a.name,
					Path:    path,
				})
			}
			continue
		}

		newJSON, _ = sjson.SetBytesOptions(newJSON, path, a.placeholder, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})
	}

	return newJSON, errors
}
