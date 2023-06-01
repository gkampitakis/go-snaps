package match

import "errors"

var missingPath = errors.New("path does not exist")

type JSONMatcher interface {
	JSON([]byte) ([]byte, []MatcherError)
}

type StructMatcher interface {
	Struct(interface{}) (interface{}, []MatcherError)
}

// internal Error struct returned from Matchers
type MatcherError struct {
	Reason  error
	Matcher string
	Path    string
}
