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

func Custom(path string, callback func(val interface{}) (interface{}, error)) *customMatcher {
	return &customMatcher{
		path:     path,
		callback: callback,
		name:     "Custom",
	}
}

func (c *customMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	r := gjson.GetBytes(s, c.path)
	if !r.Exists() {
		if c.errOnMissingPath {
			return s, []MatcherError{
				{
					Reason:  fmt.Sprintf("path %s does not exist", c.path),
					Matcher: c.name,
					Path:    c.path,
				},
			}
		}
		return s, nil
	}

	value, err := c.callback(r.Value())
	if err != nil {
		return s, []MatcherError{
			{
				Reason:  err.Error(),
				Matcher: c.name,
				Path:    c.path,
			},
		}
	}

	s, _ = sjson.SetBytesOptions(s, c.path, value, &sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	})
	return s, nil
}