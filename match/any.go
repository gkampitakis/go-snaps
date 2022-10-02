package match

type anyMatcher struct {
	paths       []string
	placeholder string
}

func Any(paths ...string) *anyMatcher {
	return &anyMatcher{
		paths:       paths,
		placeholder: "<ignore value>",
	}
}

func (a anyMatcher) JSONMatcher() JSONMatcher {
	return func(s []byte) ([]byte, string) {
		newJSON := s
		for _, path := range a.paths {
			// FIXME: error handling
			newJSON, _ = setJSON(newJSON, path, a.placeholder)
		}
		return newJSON, ""
	}
}
