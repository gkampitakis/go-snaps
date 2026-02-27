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

func expandArrayPaths(jsonInput []byte, path string) ([]string, error) {
	if path == "#." {
		r := gjson.ParseBytes(jsonInput)
		if !r.IsArray() {
			return []string{}, nil
		}

		paths := make([]string, 0, len(r.Array()))
		for i := range r.Array() {
			paths = append(paths, strconv.Itoa(i))
		}
		return paths, nil
	}

	// split on the first intermediate #, if present
	pathToArray, restOfPath, hasArrayPlaceholder := strings.Cut(path, ".#.")

	// if there is no intermediate placeholder, check for (and cut) a terminal one
	if !hasArrayPlaceholder {
		pathToArray, hasArrayPlaceholder = strings.CutSuffix(path, ".#")
	}

	// if there are no array placeholders in the path, just return it
	if !hasArrayPlaceholder {
		return []string{path}, nil
	}

	r := gjson.GetBytes(jsonInput, pathToArray)
	if !r.Exists() {
		return []string{}, errPathNotFound
	}
	// skip properties that are not arrays
	if !r.IsArray() {
		return []string{}, nil
	}

	// if property exists and is actually an array, build out the path to each item
	// within that array
	paths := make([]string, 0, len(r.Array()))

	for i := range r.Array() {
		static := pathToArray + "." + strconv.Itoa(i)

		if restOfPath != "" {
			static += "." + restOfPath
		}
		nestedPaths, err := expandArrayPaths(jsonInput, static)
		if err != nil {
			return nil, err
		}

		paths = append(paths, nestedPaths...)
	}

	return paths, nil
}
