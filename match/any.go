package match

import (
	internal_yaml "github.com/gkampitakis/go-snaps/match/internal/yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths            []string
	placeholder      any
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
func (a *anyMatcher) Placeholder(p any) *anyMatcher {
	a.placeholder = p
	return a
}

// ErrOnMissingPath determines if matcher will fail in case of trying to access a json path
// that doesn't exist
func (a *anyMatcher) ErrOnMissingPath(e bool) *anyMatcher {
	a.errOnMissingPath = e
	return a
}

// YAML is intended to be called internally on snaps.MatchYAML for applying Any matchers
func (a anyMatcher) YAML(b []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	f, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return b, []MatcherError{{
			Reason:  err,
			Matcher: a.name,
			Path:    "*",
		}}
	}

	for _, p := range a.paths {
		path, _, exists, err := internal_yaml.Get(f, p)
		if err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    p,
			})

			continue
		}
		if !exists {
			if a.errOnMissingPath {
				errs = append(errs, MatcherError{
					Reason:  errPathNotFound,
					Matcher: a.name,
					Path:    p,
				})
			}

			continue
		}

		if err := internal_yaml.Update(f, path, a.placeholder); err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    p,
			})
			continue
		}
	}

	return []byte(f.String()), errs
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Any matchers
func (a anyMatcher) JSON(b []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	json := b
	for _, path := range a.paths {
		r := gjson.GetBytes(json, path)
		if !r.Exists() {
			if a.errOnMissingPath {
				errs = append(errs, MatcherError{
					Reason:  errPathNotFound,
					Matcher: a.name,
					Path:    path,
				})
			}

			continue
		}

		j, err := sjson.SetBytesOptions(json, path, a.placeholder, setJsonOptions)
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
