package match

import (
	"bytes"
)

const (
	DELIMITER = '.'
	ESCAPE    = '\\'
	NUMBER    = '#'
)

// Still need for support
//  -------
// # returns length
// * ? wildcards

func parsePath(ss string) []string {
	parts := []string{}
	part := bytes.Buffer{}

	for i := 0; i < len(ss); i++ {
		if ss[i] == ESCAPE {
			i++
			// if we are not escaping a special char then we keep the /
			if ss[i] != DELIMITER {
				part.WriteByte(ESCAPE)
			}

			part.WriteByte(ss[i])
			continue
		}

		if ss[i] != DELIMITER {
			part.WriteByte(ss[i])
			continue
		}

		if part.Len() == 0 {
			continue
		}

		parts = append(parts, part.String())
		part.Reset()
	}

	if part.Len() > 0 {
		parts = append(parts, part.String())
	}

	return parts
}
