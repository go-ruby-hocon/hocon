// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-hocon/hocon authors

package hocon

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sample = `
	name = app
	port = 8080
	ratio = 0.5
	debug = true
	tags = [a, b, c]
	db = { host = localhost, port = 5432 }
	timeout = 30s
	cache = 10MB
`

func TestParseAndGetters(t *testing.T) {
	c, err := Parse(sample)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c.GetString("name"); v != "app" {
		t.Errorf("name = %q", v)
	}
	if v, _ := c.GetInt("port"); v != 8080 {
		t.Errorf("port = %d", v)
	}
	if v, _ := c.GetDouble("ratio"); v != 0.5 {
		t.Errorf("ratio = %v", v)
	}
	if v, _ := c.GetBoolean("debug"); v != true {
		t.Errorf("debug = %v", v)
	}
	if v, _ := c.GetList("tags"); len(v) != 3 {
		t.Errorf("tags len = %d", len(v))
	}
	if !c.HasPath("db.host") {
		t.Error("db.host should exist")
	}
	if c.Root().Type() != ObjectType {
		t.Error("root should be object")
	}
	if v, _ := c.GetValue("port"); v.Type() != NumberType {
		t.Errorf("port value type = %v", v.Type())
	}
	sub, err := c.GetConfig("db")
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := sub.GetInt("port"); v != 5432 {
		t.Errorf("db.port = %d", v)
	}
	if d, _ := c.GetDuration("timeout"); d != 30*time.Second {
		t.Errorf("timeout = %v", d)
	}
	if b, _ := c.GetBytes("cache"); b != 10_000_000 {
		t.Errorf("cache = %d", b)
	}
}

func TestParseError(t *testing.T) {
	if _, err := Parse(`a = ${x`); err == nil {
		t.Error("expected parse error")
	}
}

func TestGetConfigError(t *testing.T) {
	c, _ := Parse(`a = 1`)
	if _, err := c.GetConfig("missing"); err == nil {
		t.Error("expected error")
	}
}

func TestConfigFactory(t *testing.T) {
	c, err := ConfigFactory.ParseString(`a = 1`)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c.GetInt("a"); v != 1 {
		t.Errorf("a = %d", v)
	}
}

func TestParseFileSeam(t *testing.T) {
	reader := func(path string) (string, error) {
		if path == "app.conf" {
			return "a = 42", nil
		}
		return "", errors.New("not found")
	}
	c, err := ConfigFactory.ParseFile("app.conf", reader)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c.GetInt("a"); v != 42 {
		t.Errorf("a = %d", v)
	}
	// reader error path
	if _, err := ConfigFactory.ParseFile("missing.conf", reader); err == nil {
		t.Error("expected reader error")
	}
}

func TestParseFileDefaultReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.conf")
	if err := os.WriteFile(path, []byte("k = 7"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := ConfigFactory.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c.GetInt("k"); v != 7 {
		t.Errorf("k = %d", v)
	}
	if _, err := ConfigFactory.ParseFile(filepath.Join(dir, "nope.conf")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestWithFallback(t *testing.T) {
	base, _ := Parse("a = 1\nb = 2")
	over, _ := Parse("b = 20\nc = 30")
	m := over.WithFallback(base)
	if v, _ := m.GetInt("a"); v != 1 {
		t.Errorf("a = %d", v)
	}
	if v, _ := m.GetInt("b"); v != 20 {
		t.Errorf("b = %d", v)
	}
	if v, _ := m.GetInt("c"); v != 30 {
		t.Errorf("c = %d", v)
	}
}

func TestRenderOptions(t *testing.T) {
	c, _ := Parse("a = 1\nb = hi")
	opts := DefaultRenderOptions()
	hoconOut := c.Render(opts)
	if !strings.Contains(hoconOut, "=") {
		t.Errorf("HOCON render: %s", hoconOut)
	}
	jsonOpts := DefaultRenderOptions().SetJSON(true).SetIndent("    ")
	jsonOut := c.Render(jsonOpts)
	if !strings.Contains(jsonOut, ":") || !strings.Contains(jsonOut, "    ") {
		t.Errorf("JSON render: %s", jsonOut)
	}
	// round trip
	c2, err := Parse(jsonOut)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := c2.GetInt("a"); v != 1 {
		t.Errorf("round-trip a = %d", v)
	}
}

func TestConfigValueFactory(t *testing.T) {
	if ConfigValueFactory.FromString("s").Type() != StringType {
		t.Error("FromString")
	}
	if ConfigValueFactory.FromInt(3).Type() != NumberType {
		t.Error("FromInt")
	}
	if ConfigValueFactory.FromDouble(1.5).Type() != NumberType {
		t.Error("FromDouble")
	}
	if ConfigValueFactory.FromBoolean(true).Type() != BooleanType {
		t.Error("FromBoolean")
	}
	if ConfigValueFactory.FromNull().Type() != NullType {
		t.Error("FromNull")
	}
}

func TestFromAnyRef(t *testing.T) {
	f := ConfigValueFactory
	if f.FromAnyRef(nil).Type() != NullType {
		t.Error("nil")
	}
	if f.FromAnyRef(true).Type() != BooleanType {
		t.Error("bool")
	}
	if f.FromAnyRef(5).Type() != NumberType {
		t.Error("int")
	}
	if f.FromAnyRef(int64(5)).Type() != NumberType {
		t.Error("int64")
	}
	if f.FromAnyRef(2.5).Type() != NumberType {
		t.Error("float64")
	}
	if f.FromAnyRef("s").Type() != StringType {
		t.Error("string")
	}
	// passthrough of an existing ConfigValue
	pre := ConfigValueFactory.FromString("x")
	if f.FromAnyRef(pre) != pre {
		t.Error("passthrough")
	}
	// list
	lst := f.FromAnyRef([]any{1, "two", true})
	if lst.Type() != ArrayType || len(lst.Elements()) != 3 {
		t.Errorf("list = %v", lst)
	}
	// map (nested)
	obj := f.FromAnyRef(map[string]any{"k": 1, "nested": map[string]any{"q": 2}})
	if obj.Type() != ObjectType {
		t.Fatal("map should be object")
	}
	fields := obj.Fields()
	if fields["k"].Type() != NumberType {
		t.Errorf("k type = %v", fields["k"].Type())
	}
	if fields["nested"].Type() != ObjectType {
		t.Errorf("nested type = %v", fields["nested"].Type())
	}
	// unknown type -> stringified
	if v := f.FromAnyRef(int32(7)); v.Type() != StringType {
		t.Errorf("int32 fallback type = %v", v.Type())
	}
}
