// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-hocon/hocon authors

package hocon

import (
	"fmt"
	"os"
	"time"

	gohocon "github.com/go-hocon/hocon"
)

// ConfigValue is a single HOCON value, re-exported from the engine so callers
// depend on a single import.
type ConfigValue = gohocon.ConfigValue

// ConfigValueType classifies a [ConfigValue], re-exported from the engine.
type ConfigValueType = gohocon.ConfigValueType

// Value type constants, re-exported from the engine.
const (
	NullType    = gohocon.NullType
	BooleanType = gohocon.BooleanType
	NumberType  = gohocon.NumberType
	StringType  = gohocon.StringType
	ArrayType   = gohocon.ArrayType
	ObjectType  = gohocon.ObjectType
)

// Config mirrors Ruby's Hocon::Config: a resolved configuration tree with
// typed, path-addressed accessors.
type Config struct {
	c *gohocon.Config
}

// Parse mirrors Ruby's Hocon.parse: it parses a HOCON string into a [Config].
func Parse(s string) (*Config, error) {
	c, err := gohocon.Parse(s)
	if err != nil {
		return nil, err
	}
	return &Config{c: c}, nil
}

// FileReader loads the text of a file for [ConfigFactory.ParseFile]. It is an
// injectable seam so tests need not touch disk.
type FileReader func(path string) (string, error)

func osFileReader(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// factory mirrors Ruby's Hocon::ConfigFactory.
type factory struct{}

// ConfigFactory is the entry object mirroring Ruby's Hocon::ConfigFactory.
var ConfigFactory factory

// ParseString mirrors ConfigFactory.parse_string.
func (factory) ParseString(s string) (*Config, error) { return Parse(s) }

// ParseFile mirrors ConfigFactory.parse_file. An optional [FileReader] overrides
// the default (disk) loader, letting tests supply content in memory.
func (factory) ParseFile(path string, reader ...FileReader) (*Config, error) {
	read := osFileReader
	if len(reader) > 0 {
		read = reader[0]
	}
	src, err := read(path)
	if err != nil {
		return nil, err
	}
	return Parse(src)
}

// GetValue returns the raw [ConfigValue] at a dotted path. As in Ruby's hocon,
// each `.` starts a new object level; a key that contains a literal dot (written
// quoted, e.g. `"a.b" = 1`) is reached with [Config.GetValuePath] instead.
func (c *Config) GetValue(path string) (*ConfigValue, error) { return c.c.GetValue(path) }

// GetValuePath returns the raw [ConfigValue] at a path given as explicit
// segments, bypassing dotted-path parsing. This mirrors accessing a literal-dot
// key in Ruby's hocon by quoting the segment: for `"a.b" = 1`,
// GetValuePath("a.b") returns 1 whereas GetValue("a.b") looks for nested a.b.
func (c *Config) GetValuePath(segments ...string) (*ConfigValue, error) {
	return c.c.GetValuePath(segments...)
}

// HasPathSegments mirrors Config#has_path? for an explicit, non-dotted path.
func (c *Config) HasPathSegments(segments ...string) bool {
	return c.c.HasPathSegments(segments...)
}

// GetString mirrors Config#get_string.
func (c *Config) GetString(path string) (string, error) { return c.c.GetString(path) }

// GetInt mirrors Config#get_int.
func (c *Config) GetInt(path string) (int64, error) { return c.c.GetInt(path) }

// GetBoolean mirrors Config#get_boolean.
func (c *Config) GetBoolean(path string) (bool, error) { return c.c.GetBool(path) }

// GetDouble mirrors Config#get_double.
func (c *Config) GetDouble(path string) (float64, error) { return c.c.GetFloat(path) }

// GetList mirrors Config#get_list.
func (c *Config) GetList(path string) ([]*ConfigValue, error) { return c.c.GetList(path) }

// GetConfig mirrors Config#get_config: the object at path as a nested [Config].
func (c *Config) GetConfig(path string) (*Config, error) {
	sub, err := c.c.GetObject(path)
	if err != nil {
		return nil, err
	}
	return &Config{c: sub}, nil
}

// GetDuration mirrors Config#get_duration.
func (c *Config) GetDuration(path string) (time.Duration, error) { return c.c.GetDuration(path) }

// GetBytes mirrors Config#get_bytes.
func (c *Config) GetBytes(path string) (int64, error) { return c.c.GetBytes(path) }

// HasPath mirrors Config#has_path?.
func (c *Config) HasPath(path string) bool { return c.c.HasPath(path) }

// Root mirrors Config#root: the root object value.
func (c *Config) Root() *ConfigValue { return c.c.Root() }

// WithFallback mirrors Config#with_fallback: values missing from c are filled
// from fallback (objects deep-merge, c winning).
func (c *Config) WithFallback(fallback *Config) *Config {
	return &Config{c: c.c.WithFallback(fallback.c)}
}

// ConfigRenderOptions mirrors Hocon::ConfigRenderOptions.
type ConfigRenderOptions struct {
	// JSON renders strict JSON instead of HOCON.
	JSON bool
	// Indent is the per-level indent (empty means the engine default).
	Indent string
}

// DefaultRenderOptions returns the default (HOCON) render options.
func DefaultRenderOptions() ConfigRenderOptions { return ConfigRenderOptions{} }

// SetJSON returns a copy with JSON toggled (mirrors set_json).
func (o ConfigRenderOptions) SetJSON(b bool) ConfigRenderOptions {
	o.JSON = b
	return o
}

// SetIndent returns a copy with the indent set.
func (o ConfigRenderOptions) SetIndent(indent string) ConfigRenderOptions {
	o.Indent = indent
	return o
}

// Render serialises the config using the given options (mirrors
// config.root.render(options)).
func (c *Config) Render(opts ConfigRenderOptions) string {
	return c.c.Render(gohocon.RenderOptions{JSON: opts.JSON, Indent: opts.Indent})
}

// valueFactory mirrors Ruby's Hocon::ConfigValueFactory.
type valueFactory struct{}

// ConfigValueFactory is the entry object mirroring Hocon::ConfigValueFactory.
var ConfigValueFactory valueFactory

// FromString wraps a string.
func (valueFactory) FromString(s string) *ConfigValue { return gohocon.NewString(s) }

// FromInt wraps an integer.
func (valueFactory) FromInt(n int64) *ConfigValue { return gohocon.NewNumber(float64(n)) }

// FromDouble wraps a float.
func (valueFactory) FromDouble(f float64) *ConfigValue { return gohocon.NewNumber(f) }

// FromBoolean wraps a boolean.
func (valueFactory) FromBoolean(b bool) *ConfigValue { return gohocon.NewBool(b) }

// FromNull returns a null value.
func (valueFactory) FromNull() *ConfigValue { return gohocon.NewNull() }

// FromAnyRef mirrors ConfigValueFactory.from_any_ref: it converts an arbitrary
// Go value (nil, bool, integer, float, string, []any, map[string]any or an
// existing *ConfigValue) into a [ConfigValue]. Any other type is stringified.
func (vf valueFactory) FromAnyRef(v any) *ConfigValue {
	switch x := v.(type) {
	case nil:
		return gohocon.NewNull()
	case *ConfigValue:
		return x
	case bool:
		return gohocon.NewBool(x)
	case int:
		return gohocon.NewNumber(float64(x))
	case int64:
		return gohocon.NewNumber(float64(x))
	case float64:
		return gohocon.NewNumber(x)
	case string:
		return gohocon.NewString(x)
	case []any:
		elems := make([]*ConfigValue, len(x))
		for i, e := range x {
			elems[i] = vf.FromAnyRef(e)
		}
		return gohocon.NewArray(elems...)
	case map[string]any:
		entries := make(map[string]*ConfigValue, len(x))
		for k, e := range x {
			entries[k] = vf.FromAnyRef(e)
		}
		return gohocon.NewObjectOf(entries)
	default:
		return gohocon.NewString(fmt.Sprint(x))
	}
}
