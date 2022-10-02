package match

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths       []string
	placeholder string
}

// NOTE: order matters
// TODO: add comments on public fns
func Any(paths ...string) *anyMatcher {
	return &anyMatcher{
		paths:       paths,
		placeholder: "<Any value>",
	}
}

func (a *anyMatcher) Placeholder(p string) *anyMatcher {
	a.placeholder = p
	return a
}

// NOTE: we need to finalize the JSONMatcher return values
// Finalize the any functionality
// the matcher needs to be extensible. If not it's done
func (a anyMatcher) JSONMatcher() JSONMatcher {
	return func(s []byte) ([]byte, string) {
		newJSON := s
		for _, path := range a.paths {
			r := gjson.GetBytes(newJSON, path)
			if !r.Exists() {
				continue
			}

			newJSON, _ = sjson.SetBytesOptions(newJSON, path, a.placeholder, &sjson.Options{
				Optimistic:     true,
				ReplaceInPlace: true,
			})
		}
		return newJSON, ""
	}
}
