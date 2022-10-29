package match

import (
	"bytes"
)

const (
	DELIMITER = '.'
	ESCAPE    = '\\'
)

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func parsePath(ss string) []string {
	parts := []string{}
	part := bytes.Buffer{}

	for i := 0; i < len(ss); i++ {
		if ss[i] == ESCAPE {
			continue
		}

		if ss[i] != DELIMITER {
			part.WriteByte(ss[i])
			continue
		}

		if i != 0 && ss[i-1] == ESCAPE {
			part.WriteByte(DELIMITER)
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
