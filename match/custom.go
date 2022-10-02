package match

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func Custom(path string, callback func(val interface{}) ([]byte, error)) JSONMatcher {
	return func(s []byte) ([]byte, string) {
		r := gjson.GetBytes(s, path)
		if !r.Exists() {
			return s, ""
		}

		jsonValue, err := callback(r.Value())
		if err != nil {
			return nil, err.Error()
		}

		s, _ = sjson.SetBytesOptions(s, path, jsonValue, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})

		return s, ""
	}
}
