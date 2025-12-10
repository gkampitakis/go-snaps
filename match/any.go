package match

import (
	"bytes"

	"github.com/gkampitakis/go-snaps/match/internal/yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths                  []string
	placeholder            any
	errOnMissingPath       bool
	handleNestedJSONArrays bool
	name                   string
}

func (a *anyMatcher) matcherError(err error, path string) MatcherError {
	return MatcherError{
		Reason:  err,
		Matcher: a.name,
		Path:    path,
	}
}

/*
Any matcher acts as a placeholder for any value

It replaces any targeted path with a placeholder string

	Any("user.name", "user.email")
	// or for yaml
	Any("$.user.name", "$.user.email")
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

// ErrOnMissingPath determines if matcher will fail in case of trying to access a path
// that doesn't exist
func (a *anyMatcher) ErrOnMissingPath(e bool) *anyMatcher {
	a.errOnMissingPath = e
	return a
}

// HandleNestedJSONArrays enables support for nested arrays when using Any matcher in JSON
// E.g. results.#.packages.#.vulnerabilities
//
// see issue https://github.com/gkampitakis/go-snaps/issues/84 for details.
// Disabled by default for backward compatibility and will be enabled by default if it doesn't introduce any issues in future.
func (a *anyMatcher) HandleNestedJSONArrays() *anyMatcher {
	a.handleNestedJSONArrays = true
	return a
}

// YAML is intended to be called internally on snaps.MatchYAML for applying Any matchers
func (a anyMatcher) YAML(b []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	f, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return b, []MatcherError{a.matcherError(err, "*")}
	}

	for _, p := range a.paths {
		path, _, exists, err := yaml.Get(f, p)
		if err != nil {
			errs = append(errs, a.matcherError(err, p))

			continue
		}
		if !exists {
			if a.errOnMissingPath {
				errs = append(errs, a.matcherError(errPathNotFound, p))
			}

			continue
		}

		if err := yaml.Update(f, path, a.placeholder); err != nil {
			errs = append(errs, a.matcherError(err, p))

			continue
		}
	}

	return yaml.MarshalFile(f, bytes.HasSuffix(b, []byte("\n"))), errs
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Any matchers
func (a anyMatcher) JSON(b []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	json := b
	for _, path := range a.paths {
		if a.handleNestedJSONArrays {
			for _, ap := range expandArrayPaths(json, path) {
				j, err := a.processPath(json, ap)

				if err != nil {
					errs = append(errs, a.matcherError(err, path))
					continue
				}

				json = j
			}
		} else {
			j, err := a.processPath(json, path)
			if err != nil {
				errs = append(errs, a.matcherError(err, path))

				continue
			}

			json = j
		}
	}

	return json, errs
}

func (a anyMatcher) processPath(json []byte, path string) ([]byte, error) {
	r := gjson.GetBytes(json, path)
	if !r.Exists() {
		if a.errOnMissingPath {
			return nil, errPathNotFound
		}

		return json, nil
	}

	j, err := sjson.SetBytesOptions(json, path, a.placeholder, setJSONOptions)
	if err != nil {
		return nil, err
	}

	return j, nil
}
