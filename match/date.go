package match

import (
	"fmt"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type dateMatcher struct {
	path        string
	layout      string
	placeholder string
}

func Date(path string) *dateMatcher {
	return &dateMatcher{
		path:        path,
		placeholder: "<Any Date>",
		layout:      time.RFC3339,
	}
}

func (d *dateMatcher) Placeholder(p string) *dateMatcher {
	d.placeholder = p
	return d
}

func (d *dateMatcher) Format(l string) *dateMatcher {
	d.layout = l
	return d
}

func (d *dateMatcher) JSONMatcher() JSONMatcher {
	return func(json []byte) ([]byte, string) {
		r := gjson.GetBytes(json, d.path)

		if !d.validate(r.String()) {
			return nil, fmt.Sprintf("expected %s date but got %s", d.layout, r.String())
		}

		json, _ = sjson.SetBytes(json, d.layout, d.placeholder)
		return json, ""
	}
}

// internal
func (d *dateMatcher) validate(s string) bool {
	_, err := time.Parse(d.layout, s)
	return err == nil
}
