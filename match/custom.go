package match

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type customMatcher struct {
	name             string
	path             string
	errOnMissingPath bool
	callback         func(val interface{}) (interface{}, error)
}

/*
Custom matcher provides a matcher where you can define your own validation and placeholder

	match.Custom("user.age", func(val interface{})) (interface{}, error) {
		age, ok := val.(float64)
		if !ok {
				return nil, fmt.Errorf("expected number but got %T", val)
		}

		return "some number>", nil
	}
*/
func Custom(path string, callback func(val interface{}) (interface{}, error)) *customMatcher {
	return &customMatcher{
		path:     path,
		callback: callback,
		name:     "Custom",
	}
}

// ErrOnMissingPath will make the fail in case a path accessed doesn't exist
func (c *customMatcher) ErrOnMissingPath() *customMatcher {
	c.errOnMissingPath = true
	return c
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Custom matcher
func (c *customMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	r := gjson.GetBytes(s, c.path)
	if !r.Exists() && c.errOnMissingPath {
		return s, []MatcherError{
			{
				Reason:  fmt.Errorf("path %s does not exist", c.path),
				Matcher: c.name,
				Path:    c.path,
			},
		}
	}

	value, err := c.callback(r.Value())
	if err != nil {
		return s, []MatcherError{
			{
				Reason:  err,
				Matcher: c.name,
				Path:    c.path,
			},
		}
	}

	s, err = sjson.SetBytesOptions(s, c.path, value, &sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	})
	if err != nil {
		return s, []MatcherError{
			{
				Reason:  err,
				Matcher: c.name,
				Path:    c.path,
			},
		}
	}
	return s, nil
}
