package match

type JSONMatcher interface {
	JSON([]byte) ([]byte, []MatcherError)
}

type MatcherError struct {
	Reason  string
	Matcher string
	Path    string
}
