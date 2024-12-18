package match

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var errPathNotFound = errors.New("path does not exist")

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

func (a anyMatcher) YAML(s []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	f, err := parser.ParseBytes(s, parser.ParseComments)
	if err != nil {
		return s, []MatcherError{{
			Reason:  err,
			Matcher: a.name,
			Path:    "*",
		}}
	}

	for _, path := range a.paths {
		p, err := yaml.PathString(path)
		if err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    path,
			})

			continue
		}

		if a.errOnMissingPath {
			if _, err = p.FilterFile(f); err != nil && errors.Is(err, yaml.ErrNotFoundNode) {
				errs = append(errs, MatcherError{
					Reason:  errPathNotFound,
					Matcher: a.name,
					Path:    path,
				})
				continue
			}
		}

		err = p.ReplaceWithReader(f, strings.NewReader(fmt.Sprint(a.placeholder)))
		if err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    path,
			})
			continue
		}
	}

	return []byte(f.String()), errs
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
					Reason:  errPathNotFound,
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
