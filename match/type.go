package match

import (
	"bytes"
	"fmt"

	"github.com/gkampitakis/go-snaps/match/internal/yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type typeMatcher[ExpectedType any] struct {
	paths            []string
	errOnMissingPath bool
	name             string
	expectedType     any
}

func (t *typeMatcher[ExpectedType]) matcherError(err error, path string) MatcherError {
	return MatcherError{
		Reason:  err,
		Matcher: t.name,
		Path:    path,
	}
}

/*
Type matcher evaluates types that are passed in a snapshot

It replaces any targeted path with placeholder in the form of `<Type:ExpectedType>`

	match.Type[string]("user.info", "user.age")
	// or for yaml
	match.Type[string]("$.user.info", "$.user.age")
*/
func Type[ExpectedType any](paths ...string) *typeMatcher[ExpectedType] {
	return &typeMatcher[ExpectedType]{
		paths:            paths,
		errOnMissingPath: true,
		name:             "Type",
		expectedType:     *new(ExpectedType),
	}
}

// ErrOnMissingPath determines if matcher will fail in case of trying to access a path
// that doesn't exist
func (t *typeMatcher[T]) ErrOnMissingPath(e bool) *typeMatcher[T] {
	t.errOnMissingPath = e
	return t
}

// YAML is intended to be called internally on snaps.MatchJSON for applying Type matchers
func (t typeMatcher[ExpectedType]) YAML(b []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	f, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return b, []MatcherError{t.matcherError(err, "*")}
	}

	for _, p := range t.paths {
		path, node, exists, err := yaml.Get(f, p)
		if err != nil {
			errs = append(errs, t.matcherError(err, p))

			continue
		}
		if !exists {
			if t.errOnMissingPath {
				errs = append(errs, t.matcherError(errPathNotFound, p))
			}

			continue
		}

		value, err := yaml.GetValue(node)
		if err != nil {
			errs = append(errs, t.matcherError(err, p))

			continue
		}

		if err := typeCheck[ExpectedType](value); err != nil {
			errs = append(errs, t.matcherError(err, p))

			continue
		}

		if err := yaml.Update(f, path, typePlaceholder(value)); err != nil {
			errs = append(errs, t.matcherError(err, p))

			continue
		}
	}

	return yaml.MarshalFile(f, bytes.HasSuffix(b, []byte("\n"))), errs
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Type matchers
func (t typeMatcher[ExpectedType]) JSON(b []byte) ([]byte, []MatcherError) {
	var errs []MatcherError
	json := b

	for _, path := range t.paths {
		r := gjson.GetBytes(json, path)
		if !r.Exists() {
			if t.errOnMissingPath {
				errs = append(errs, t.matcherError(errPathNotFound, path))
			}

			continue
		}

		if err := typeCheck[ExpectedType](r.Value()); err != nil {
			errs = append(errs, t.matcherError(err, path))

			continue
		}

		j, err := sjson.SetBytesOptions(
			json,
			path,
			typePlaceholder(r.Value()),
			setJSONOptions,
		)
		if err != nil {
			errs = append(errs, t.matcherError(err, path))

			continue
		}

		json = j
	}

	return json, errs
}

func typeCheck[ExpectedType any](value any) error {
	if _, ok := value.(ExpectedType); !ok {
		return fmt.Errorf("expected type %T, received %T", *new(ExpectedType), value)
	}

	return nil
}

func typePlaceholder(value any) string {
	return fmt.Sprintf("<Type:%T>", value)
}
