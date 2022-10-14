package match

type JSONMatcher interface {
	JSON([]byte) ([]byte, []MatcherError)
}

type MatcherError struct {
	Reason  error
	Matcher string
	Path    string
}
