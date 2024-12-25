package match

import (
	"bytes"

	"github.com/gkampitakis/go-snaps/match/internal/yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type customMatcher struct {
	callback         func(val any) (any, error)
	errOnMissingPath bool
	name             string
	path             string
}

func (c *customMatcher) matcherError(err error) []MatcherError {
	return []MatcherError{{
		Reason:  err,
		Matcher: c.name,
		Path:    c.path,
	}}
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
	 bool // for YAML booleans
	 float64 // for YAML float numbers
	 uint64 // for YAML integer numbers
	 string // for YAML string literals
	 nil // for YAML null
	 map[string]any // for YAML objects
	 []any // for YAML arrays
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
		return nil, c.matcherError(err)
	}

	path, node, exists, err := yaml.Get(f, c.path)
	if err != nil {
		return nil, c.matcherError(err)
	}
	if !exists {
		if c.errOnMissingPath {
			return nil, c.matcherError(errPathNotFound)
		}

		return b, nil
	}

	value, err := yaml.GetValue(node)
	if err != nil {
		return nil, c.matcherError(err)
	}

	result, err := c.callback(value)
	if err != nil {
		return nil, c.matcherError(err)
	}

	if err := yaml.Update(f, path, result); err != nil {
		return nil, c.matcherError(err)
	}

	return yaml.MarshalFile(f, bytes.HasSuffix(b, []byte("\n"))), nil
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Custom matcher
func (c *customMatcher) JSON(b []byte) ([]byte, []MatcherError) {
	r := gjson.GetBytes(b, c.path)
	if !r.Exists() {
		if c.errOnMissingPath {
			return nil, c.matcherError(errPathNotFound)
		}

		return b, nil
	}

	value, err := c.callback(r.Value())
	if err != nil {
		return nil, c.matcherError(err)
	}

	b, err = sjson.SetBytesOptions(b, c.path, value, setJSONOptions)
	if err != nil {
		return nil, c.matcherError(err)
	}

	return b, nil
}
