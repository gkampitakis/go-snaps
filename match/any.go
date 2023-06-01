package match

import (
	"errors"
	"reflect"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type anyMatcher struct {
	paths            []string
	placeholder      interface{}
	errOnMissingPath bool
	name             string
}

/*
Any matcher acts as a placeholder for any value

It replaces any targeted path with a placeholder string

	Any("user.name")
	// or with multiple paths
	Any("user.name", "user.email")
*/
func Any(paths ...string) *anyMatcher {
	return &anyMatcher{
		errOnMissingPath: true,
		placeholder:      "<Any value>",
		paths:            paths,
		name:             "Any",
	}
}

// Placeholder allows to define the placeholder value for Any matcher
func (a *anyMatcher) Placeholder(p interface{}) *anyMatcher {
	a.placeholder = p
	return a
}

// ErrOnMissingPath determines if matcher will fail in case of trying to access a json path
// that doesn't exist
func (a *anyMatcher) ErrOnMissingPath(e bool) *anyMatcher {
	a.errOnMissingPath = e
	return a
}

// JSON is intended to be called internally on snaps.MatchJSON for applying Any matchers
func (a anyMatcher) JSON(s []byte) ([]byte, []MatcherError) {
	var errs []MatcherError

	json := s
	for _, path := range a.paths {
		r := gjson.GetBytes(json, path)
		if !r.Exists() {
			if a.errOnMissingPath {
				errs = append(errs, MatcherError{
					Reason:  missingPath,
					Matcher: a.name,
					Path:    path,
				})
			}
			continue
		}

		j, err := sjson.SetBytesOptions(json, path, a.placeholder, &sjson.Options{
			Optimistic:     true,
			ReplaceInPlace: true,
		})
		if err != nil {
			errs = append(errs, MatcherError{
				Reason:  err,
				Matcher: a.name,
				Path:    path,
			})

			continue
		}

		json = j
	}

	return json, errs
}

// TODO: we need to access unexported fields ?
// TODO: we need to set data
// TODO: missing path
// NOTE: now this only works with pointer, can we make it work with pointer and struct value?

func (a anyMatcher) Struct(s interface{}) (interface{}, []MatcherError) {
	var errs []MatcherError
	i := NewInspector(s)

	for _, path := range a.paths {
		v, err := i.SetValue(path, nil)
		if err != nil {
			if errors.Is(err, missingPath) && !a.errOnMissingPath {
				continue
			}
			errs = append(errs, MatcherError{Reason: err, Matcher: a.name, Path: path})
			continue
		}
		s = v
	}

	return s, errs
}

// First draft we want this to work only for struct data
// should fail if path doesn't exist

type StructInspector struct {
	o interface{}
}

func NewInspector(value interface{}) *StructInspector {
	return &StructInspector{
		o: value,
	}
}

func (si *StructInspector) GetValue(path string) (interface{}, error) {
	rv := reflect.ValueOf(si.o)
	parts := parsePath(path)

	for i, part := range parts {
		v := rv.FieldByName(part)
		if !v.IsValid() {
			return nil, missingPath
		}
		_ = i
		rv = v
	}

	return rv.Interface(), nil
}

func (si *StructInspector) SetValue(path string, value interface{}) (interface{}, error) {
	rv := reflect.ValueOf(si.o).Elem()
	parts := parsePath(path)

	for _, part := range parts {
		v := rv.FieldByName(part)
		if !v.IsValid() {
			return si.o, missingPath
		}
		rv = v
	}

	// set empty value of type
	// how does this behave with custom types ?
	rv.Set(reflect.New(rv.Type()).Elem())
	return si.o, nil
}
