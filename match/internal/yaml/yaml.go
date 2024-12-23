package internal

import (
	"bytes"
	"errors"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

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

func Update(f *ast.File, path *yaml.Path, value interface{}) error {
	b, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	return path.ReplaceWithReader(f, bytes.NewReader(b))
}
