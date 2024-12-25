package yaml

import (
	"bytes"
	"errors"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// GetValue returns the value of the node.
func GetValue(node ast.Node) (interface{}, error) {
	data, err := node.MarshalYAML()
	if err != nil {
		return nil, err
	}

	var value interface{}
	if err := yaml.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return value, nil
}

// Get takes an ast.File and a string representing a path
// and returns the yaml.Path, the node and a bool indicating if the node exists.
func Get(f *ast.File, p string) (*yaml.Path, ast.Node, bool, error) {
	path, err := yaml.PathString(p)
	if err != nil {
		return nil, nil, false, err
	}

	node, err := path.FilterFile(f)
	if err != nil {
		if errors.Is(err, yaml.ErrNotFoundNode) {
			return path, nil, false, nil
		}

		return path, nil, false, err
	}

	return path, node, true, nil
}

// Update marshals the value and replaces the file at the path provided with the new value.
func Update(f *ast.File, path *yaml.Path, value interface{}) error {
	b, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	return path.ReplaceWithReader(f, bytes.NewReader(b))
}

// MarshalFile returns the representation of the ast.File to a byte slice.
func MarshalFile(f *ast.File, addNewLine bool) []byte {
	docs := make([]string, 0, len(f.Docs))

	for _, doc := range f.Docs {
		docs = append(docs, doc.String())
	}

	if addNewLine {
		docs = append(docs, "")
	}

	return []byte(strings.Join(docs, "\n"))
}
