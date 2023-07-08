package snaps

import (
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestMatchInlineSnapshot(t *testing.T) {
	t.Run("should error in case of different input from inline snapshot", func(t *testing.T) {
		mockT := test.NewMockTestingT(t)

		mockT.MockError = func(a ...any) {
			test.Equal(
				t,
				"\n\x1b[38;5;52m\x1b[48;5;225m- Snapshot - 1\x1b[0m\n\x1b[38;5;22m\x1b[48;5;159m+ Received + 0\x1b[0m\n\n\x1b[48;5;225m\x1b"+
					"[38;5;52m- \x1b[0m\x1b[48;5;127m\x1b[38;5;255mdifferent \x1b[0m\x1b[48;5;225m\x1b[38;5;52mvalue\x1b[0m\n\x1b[48;5;159m\x1b"+
					"[38;5;22m+ \x1b[0m\x1b[48;5;159m\x1b[38;5;22mvalue\x1b[0m\n\n",
				a[0].(string),
			)
		}

		MatchInlineSnapshot(mockT, "value", Inline("different value"))

		test.Equal(t, 1, testEvents.items[erred])
	})
}
