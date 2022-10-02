package match

import (
	"errors"
)

var ErrPathNotExist = errors.New("path does not exist")

// NOTE: is this the best type ? somehow we need to return errors
type JSONMatcher func([]byte) ([]byte, string)
