package match

import (
	"errors"

	"github.com/tidwall/sjson"
)

var (
	errPathNotFound = errors.New("path does not exist")
	setJSONOptions  = &sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	}
)

type JSONMatcher interface {
	JSON([]byte) ([]byte, []MatcherError)
}

type YAMLMatcher interface {
	YAML([]byte) ([]byte, []MatcherError)
}

// internal Error struct returned from Matchers
type MatcherError struct {
	Reason  error
	Matcher string
	Path    string
}
