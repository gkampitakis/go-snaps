package match

import (
	"errors"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type customMatcher struct {
	callback         func(val interface{}) (interface{}, error)
	errOnMissingPath bool
	name             string
	path             string
}

type CustomCallback func(val interface{}) (interface{}, error)

/*
Custom matcher provides a matcher where you can define your own validation and placeholder

	match.Custom("user.age", func(val interface{})) (interface{}, error) {
		age, ok := val.(float64)
		if !ok {
				return nil, fmt.Errorf("expected number but got %T", val)
		}

		return "some number>", nil
	}

	The callback func value can be on of these types:
	 bool // for JSON booleans
	 float64 // for JSON numbers
	 string // for JSON string literals
	 nil // for JSON null
	 map[string]interface{} // for JSON objects
	 []interface{} // for JSON arrays
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

// JSON is intended to be called internally on snaps.MatchJSON for applying Custom matcher
func (c *customMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	r := gjson.GetBytes(s, c.path)
	if !r.Exists() {
		if c.errOnMissingPath {
			return nil, []MatcherError{{
				Reason:  errors.New("path does not exist"),
				Matcher: c.name,
				Path:    c.path,
			}}
		}

		return s, nil
	}

	value, err := c.callback(r.Value())
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	s, err = sjson.SetBytesOptions(s, c.path, value, &sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	})
	if err != nil {
		return nil, []MatcherError{{
			Reason:  err,
			Matcher: c.name,
			Path:    c.path,
		}}
	}

	return s, nil
}
