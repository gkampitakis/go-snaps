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

// gjsonExpandPath takes a gjson path, checks for nested access of arrays
// and returns all possible paths to be used for setting values
func gjsonExpandPath(data []byte, path string) ([]string, error) {
	// if no array access, return the path as is
	if !strings.Contains(path, "#") {
		return []string{path}, nil
	}

	// Get the path ending with #
	// E.g. results.#.packages.#.vulnerabilities => results.#.packages.#
	numOfEntriesPath := path[:strings.LastIndex(path, "#")+1]
	// This returns a potentially nested array of array lengths
	numOfEntries := gjson.GetBytes(data, numOfEntriesPath)
	if !numOfEntries.Exists() {
		return nil, errPathNotFound
	}

	return constructExpandedPaths(path, numOfEntries)
}

func constructExpandedPaths(path string, structure gjson.Result) ([]string, error) {
	paths := []string{}

	if structure.IsArray() {
		// More nesting to go
		for i, res := range structure.Array() {
			p, err := constructExpandedPaths(
				// Replace the first # with actual index
				strings.Replace(path, "#", strconv.Itoa(i), 1),
				res,
			)
			if err != nil {
				return nil, err
			}

			paths = append(paths, p...)
		}
	} else {
		// Otherwise assume it is a number
		if strings.Count(path, "#") != 1 {
			return nil, errors.New("programmer error: there should only be 1 # left")
		}
		for i2 := range int(structure.Int()) {
			newPath := strings.Replace(path, "#", strconv.Itoa(i2), 1)
			paths = append(paths, newPath)
		}
	}

	return paths, nil
}
