package match

type JSONMatcher interface {
	JSON([]byte) ([]byte, []MatcherError)
}

// internal Error struct returned from Matchers
type MatcherError struct {
	Reason  error
	Matcher string
	Path    string
}
