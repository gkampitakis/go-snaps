package match

import (
	"errors"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
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

func expandArrayPaths(jsonInput []byte, path string) []string {
	// split on the first intermediate #, if present
	pathToArray, restOfPath, hasArrayPlaceholder := strings.Cut(path, ".#.")

	// if there is no intermediate placeholder, check for (and cut) a terminal one
	if !hasArrayPlaceholder {
		pathToArray, hasArrayPlaceholder = strings.CutSuffix(path, ".#")
	}

	// if there are no array placeholders in the path, just return it
	if !hasArrayPlaceholder {
		return []string{path}
	}

	r := gjson.GetBytes(jsonInput, pathToArray)
	// skip properties that are not arrays
	if !r.IsArray() {
		return []string{}
	}

	// if property exists and is actually an array, build out the path to each item
	// within that array
	paths := make([]string, 0, len(r.Array()))

	for i := range r.Array() {
		static := pathToArray + "." + strconv.Itoa(i)

		if restOfPath != "" {
			static += "." + restOfPath
		}
		paths = append(paths, expandArrayPaths(jsonInput, static)...)
	}

	return paths
}
