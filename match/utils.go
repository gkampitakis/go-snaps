package match

import (
	"errors"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var ErrPathNotExist = errors.New("path does not exist")

// NOTE: is this the best type ? somehow we need to return errors
type JSONMatcher func([]byte) ([]byte, string)

func setJSON(json []byte, path, placeholder string) ([]byte, error) {
	r := gjson.GetBytes(json, path)
	if r.Exists() {
		return sjson.SetBytesOptions(json, path, placeholder, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})
	}

	return nil, ErrPathNotExist
}
