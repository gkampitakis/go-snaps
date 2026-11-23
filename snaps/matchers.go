package snaps

import (
	"fmt"
	"strings"

	"github.com/gkampitakis/go-snaps/internal/colors"
	"github.com/gkampitakis/go-snaps/match"
)

// applyMatchers runs each matcher against b in turn, collecting any errors;
// matchers that error do not modify b
func applyMatchers[M any](
	b []byte,
	matchers []M,
	apply func(M, []byte) ([]byte, []match.MatcherError),
) ([]byte, []match.MatcherError) {
	errors := []match.MatcherError{}

	for _, m := range matchers {
		out, errs := apply(m, b)
		if len(errs) > 0 {
			errors = append(errors, errs...)
			continue
		}

		b = out
	}

	return b, errors
}

func formatMatcherErrors(errs []match.MatcherError) string {
	s := strings.Builder{}

	for _, err := range errs {
		colors.Fprint(
			&s,
			colors.Red,
			fmt.Sprintf(
				"\n%smatch.%s(\"%s\") - %s",
				errorSymbol,
				err.Matcher,
				err.Path,
				err.Reason,
			),
		)
	}

	return s.String()
}
