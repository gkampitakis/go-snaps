package snaps

import (
	"fmt"

	"github.com/tidwall/pretty"
)

var defaultConfig = Config{
	snapsDir: "__snapshots__",
}

type Config struct {
	filename   string
	snapsDir   string
	extension  string
	update     *bool
	json       *JSONConfig
	serializer func(any) string
}

type JSONConfig struct {
	// Width is a max column width for single line arrays
	// Default: see defaultPrettyJSONOptions.Width for detail
	Width int
	// Indent is the nested indentation
	// Default: see defaultPrettyJSONOptions.Indent for detail
	Indent string
	// SortKeys will sort the keys alphabetically
	// Default: see defaultPrettyJSONOptions.SortKeys for detail
	SortKeys bool
}

func (j *JSONConfig) getPrettyJSONOptions() *pretty.Options {
	if j == nil {
		return defaultPrettyJSONOptions
	}
	return &pretty.Options{
		Width:    j.Width,
		Indent:   j.Indent,
		SortKeys: j.SortKeys,
	}
}

// Update determines whether to update snapshots or not
//
// It respects if running on CI.
func Update(u bool) func(*Config) {
	return func(c *Config) {
		c.update = &u
	}
}

// Specify a custom serializer function to convert the received value to a string before saving it in the snapshot file.
//
// Note: this is only used for non-structured snapshots e.g. MatchSnapshot, MatchStandaloneSnapshot, MatchInlineSnapshot.
func Serializer(s func(any) string) func(*Config) {
	return func(c *Config) {
		c.serializer = s
	}
}

// Raw is a utility function for setting serializer to fmt.Sprint
//
// For more complex custom serialization logic, use snaps.Serializer instead of snaps.Raw
func Raw() func(*Config) {
	return func(c *Config) {
		c.serializer = func(v any) string {
			return fmt.Sprint(v)
		}
	}
}

// Specify snapshot file name
//
//	default: test's filename
//
// this doesn't change the file extension see `snap.Ext`
func Filename(name string) func(*Config) {
	return func(c *Config) {
		c.filename = name
	}
}

// Specify folder name where snapshots are stored
//
//	default: __snapshots__
//
// Accepts absolute paths
func Dir(dir string) func(*Config) {
	return func(c *Config) {
		c.snapsDir = dir
	}
}

// Specify file name extension
//
// default: .snap
//
// Note: even if you specify a different extension the file still contain .snap
// e.g. if you specify .txt the file will be .snap.txt
func Ext(ext string) func(*Config) {
	return func(c *Config) {
		c.extension = ext
	}
}

// Specify json format configuration
//
// default: see defaultPrettyJSONOptions for default json config
func JSON(json JSONConfig) func(*Config) {
	return func(c *Config) {
		c.json = &json
	}
}

// Create snaps with configuration
//
//	e.g snaps.WithConfig(snaps.Filename("my_test")).MatchSnapshot(t, "hello world")
func WithConfig(args ...func(*Config)) *Config {
	s := defaultConfig

	for _, arg := range args {
		arg(&s)
	}

	return &s
}
