package match

import (
	internal_yaml "github.com/gkampitakis/go-snaps/match/internal/yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// FIXME: custom signature should be the same as the rest

type customMatcher struct {
	callback         func(val any) (any, error)
	errOnMissingPath bool
	name             string
	path             string
}

type CustomCallback func(val any) (any, error)

/*
Custom matcher allows you to bring your own validation and placeholder value.

	match.Custom("user.age", func(val any) (any, error) {
		age, ok := val.(float64)
		if !ok {
				return nil, fmt.Errorf("expected number but got %T", val)
		}

		return "some number", nil
	})

	The callback func value for JSON can be one of these types:
	 bool // for JSON booleans
	 float64 // for JSON numbers
	 string // for JSON string literals
	 nil // for JSON null
	 map[string]any // for JSON objects
	 []any // for JSON arrays

	 The callback func value for YAML can be one of these types:
	//  TODO:
*/
func Custom(path string, callback CustomCallback) *customMatcher {
	return &customMatcher{
		errOnMissingPath: true,
		callback:         callback,
		name:             "Custom",
		path:             path,
	}
}

// ErrOnMissingPath determines if Matcher will fail in case of trying to access a json path
// that doesn't exist
func (c *customMatcher) ErrOnMissingPath(e bool) *customMatcher {
	c.errOnMissingPath = e
	return c
}

// YAML is intended to be called internally on snaps.MatchYAML for applying Custom matcher
func (c *customMatcher) YAML(b []byte) ([]byte, []MatcherError) {
	f, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	path, node, exists, err := internal_yaml.Get(f, c.path)
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}
	if !exists {
		if c.errOnMissingPath {
			return nil, []MatcherError{{
				Reason:  errPathNotFound,
				Matcher: c.name,
				Path:    c.path,
			}}
		}

		return b, nil
	}

	value, err := internal_yaml.GetValue(node)
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	result, err := c.callback(value)
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	if err := internal_yaml.Update(f, path, result); err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	return []byte(f.String()), nil
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Custom matcher
func (c *customMatcher) JSON(b []byte) ([]byte, []MatcherError) {
	r := gjson.GetBytes(b, c.path)
	if !r.Exists() {
		if c.errOnMissingPath {
			return nil, []MatcherError{{
				Reason:  errPathNotFound,
				Matcher: c.name,
				Path:    c.path,
			}}
		}

		return b, nil
	}

	value, err := c.callback(r.Value())
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	b, err = sjson.SetBytesOptions(b, c.path, value, setJsonOptions)
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	return b, nil
}
