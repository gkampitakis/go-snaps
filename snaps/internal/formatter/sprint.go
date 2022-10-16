package formatter

import (
	"fmt"
	"reflect"
)

// Sprint is a convenience wrapper for fmt.Sprintf.
//
// Calling Sprint(x, y) is equivalent to
// fmt.Sprint(Formatter(x), Formatter(y)), but each operand is
// formatted with "%# v".
func Sprint(a ...interface{}) string {
	return fmt.Sprint(wrap(a, true)...)
}

func wrap(a []interface{}, force bool) []interface{} {
	w := make([]interface{}, len(a))
	for i, x := range a {
		w[i] = formatter{v: reflect.ValueOf(x), force: force}
	}
	return w
}
