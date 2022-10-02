package match

import (
	"log"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type dateMatcher struct {
	path        string
	layout      string
	placeholder string
}

// internal
func (d *dateMatcher) validate(s string) bool {
	_, err := time.Parse(d.layout, s)
	return err == nil
}

func (d *dateMatcher) Format(l string) *dateMatcher {
	d.layout = l
	return d
}

func Date(path string) *dateMatcher {
	return &dateMatcher{
		path:        path,
		placeholder: "<any date>",
	}
}

func (d *dateMatcher) JSONMatcher() JSONMatcher {
	return func(json []byte) ([]byte, string) {
		r := gjson.GetBytes(json, d.path)

		if !d.validate(r.String()) {
			log.Println("error")
		}

		json, _ = sjson.SetBytes(json, d.layout, d.placeholder)
		return json, ""
	}
}
