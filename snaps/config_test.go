package snaps

import (
	"fmt"
	"testing"

	"github.com/gkampitakis/go-snaps/internal/test"
)

func TestWithConfig(t *testing.T) {
	t.Run("returns default config when no options provided", func(t *testing.T) {
		c := WithConfig()

		test.Equal(t, "__snapshots__", c.snapsDir)
		test.Equal(t, "", c.filename)
		test.Equal(t, "", c.extension)
		test.Nil(t, c.update)
		test.Nil(t, c.json)
		test.Nil(t, c.serializer)
	})

	t.Run("Filename", func(t *testing.T) {
		c := WithConfig(Filename("my_test"))
		test.Equal(t, "my_test", c.filename)
	})

	t.Run("Dir", func(t *testing.T) {
		c := WithConfig(Dir("my_dir"))
		test.Equal(t, "my_dir", c.snapsDir)
	})

	t.Run("Ext", func(t *testing.T) {
		c := WithConfig(Ext(".txt"))
		test.Equal(t, ".txt", c.extension)
	})

	t.Run("Update", func(t *testing.T) {
		c := WithConfig(Update(true))
		test.Equal(t, true, *c.update)

		c = WithConfig(Update(false))
		test.Equal(t, false, *c.update)
	})

	t.Run("JSON", func(t *testing.T) {
		c := WithConfig(JSON(JSONConfig{SortKeys: true, Indent: "  ", Width: 80}))
		test.Equal(t, true, c.json.SortKeys)
		test.Equal(t, "  ", c.json.Indent)
		test.Equal(t, 80, c.json.Width)
	})

	t.Run("Printer", func(t *testing.T) {
		fn := func(v any) string { return fmt.Sprint(v) }
		c := WithConfig(Serializer(fn))
		test.Equal(t, "hello", c.serializer("hello"))
	})

	t.Run("multiple options are all applied", func(t *testing.T) {
		c := WithConfig(Filename("my_test"), Dir("my_dir"), Ext(".txt"), Update(true))
		test.Equal(t, "my_test", c.filename)
		test.Equal(t, "my_dir", c.snapsDir)
		test.Equal(t, ".txt", c.extension)
		test.Equal(t, true, *c.update)
	})

	t.Run("does not mutate defaultConfig", func(t *testing.T) {
		_ = WithConfig(Filename("my_test"), Dir("my_dir"), Update(true))

		test.Equal(t, "__snapshots__", defaultConfig.snapsDir)
		test.Equal(t, "", defaultConfig.filename)
		test.Nil(t, defaultConfig.update)
		test.Nil(t, defaultConfig.serializer)
	})
}

func TestTakeSnapshot(t *testing.T) {
	t.Run("falls back to pretty.Sprint when no printer set", func(t *testing.T) {
		result := defaultConfig.takeSnapshot([]any{10, "hello world"})

		test.Equal(t, "int(10)\nhello world", result)
	})

	t.Run("uses custom printer", func(t *testing.T) {
		c := WithConfig(Serializer(func(v any) string {
			return fmt.Sprintf("custom:%v", v)
		}))

		result := c.takeSnapshot([]any{"world"})

		test.Equal(t, "custom:world", result)
	})

	t.Run("calls printer once per value", func(t *testing.T) {
		calls := 0
		c := WithConfig(Serializer(func(v any) string {
			calls++
			return fmt.Sprint(v)
		}))

		value := c.takeSnapshot([]any{"a", "b", "c"})

		test.Equal(t, 3, calls)
		test.Equal(t, "a\nb\nc", value)
	})
}

func TestTakeStandaloneSnapshot(t *testing.T) {
	t.Run("falls back to pretty.Sprint when no printer set", func(t *testing.T) {
		result := defaultConfig.takeStandaloneSnapshot("hello world")

		test.Equal(t, "hello world", result)
	})

	t.Run("uses custom printer", func(t *testing.T) {
		c := WithConfig(Serializer(func(v any) string {
			return fmt.Sprintf("custom:%v", v)
		}))

		result := c.takeStandaloneSnapshot("world")

		test.Equal(t, "custom:world", result)
	})
}

func TestTakeInlineSnapshot(t *testing.T) {
	t.Run("falls back to pretty.Sprint when no printer set", func(t *testing.T) {
		result := defaultConfig.takeInlineSnapshot("hello world")

		test.Equal(t, "hello world", result)
	})

	t.Run("uses custom printer", func(t *testing.T) {
		c := WithConfig(Serializer(func(v any) string {
			return fmt.Sprintf("custom:%v", v)
		}))

		result := c.takeInlineSnapshot("world")

		test.Equal(t, "custom:world", result)
	})
}
