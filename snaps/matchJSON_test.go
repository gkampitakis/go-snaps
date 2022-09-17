package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps/internal/test"
)

func TestMatchJSON(t *testing.T) {
	t.Run("string should return not supported", func(t *testing.T) {
		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockError: func(args ...interface{}) {
				test.Equal(t, "string not implemented", args[0])
			},
		}
		MatchJSON(mockT, "string")
	})
	t.Run("[]byte should return not supported", func(t *testing.T) {
		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockError: func(args ...interface{}) {
				test.Equal(t, "[]byte not implemented", args[0])
			},
		}
		MatchJSON(mockT, []byte("[]byte"))
	})
	t.Run("should return not implemented", func(t *testing.T) {
		mockT := test.MockTestingT{
			MockHelper: func() {},
			MockErrorf: func(format string, args ...interface{}) {
				test.Equal(t, "type: %T not supported\n", format)
				test.Equal(t, 100, args[0])
			},
		}
		MatchJSON(mockT, 100)
	})
}
