// SPDX-License-Identifier: BSD-3-Clause
//
// Copyright (c) 2026, the go-ruby-hocon/hocon authors

// Package hocon is a pure-Go (no cgo) adapter that presents the Ruby `hocon`
// gem's API surface over the parser engine github.com/go-hocon/hocon.
//
// The go-hocon engine already implements HOCON semantics in full: the JSON
// superset grammar, object merging, array and value concatenation, ${...}
// substitutions with environment fallback, += self-append, include directives
// and duration / size unit suffixes. This package does not reimplement any of
// that; it wraps the engine and renames its surface to mirror the Ruby gem so a
// consumer such as go-embedded-ruby (rbgo) can expose a Ruby "Hocon" module:
//
//   - [Parse] mirrors Hocon.parse(string);
//   - [ConfigFactory] mirrors Hocon::ConfigFactory (parse_string / parse_file,
//     the latter through an injectable [FileReader] seam);
//   - [Config] mirrors Hocon::ConfigObject/Config: get_string, get_int,
//     get_boolean, get_double, get_list, get_config, has_path?, with_fallback,
//     root, get_duration and get_bytes;
//   - [ConfigValueFactory] mirrors Hocon::ConfigValueFactory.from_any_ref and
//     friends, converting native Go values into engine [ConfigValue]s;
//   - [ConfigRenderOptions] mirrors Hocon::ConfigRenderOptions, driving
//     [Config.Render] back to HOCON or JSON.
//
// The package has no dependency on any Ruby runtime: the surface is Go-typed,
// and a Ruby binding layer marshals Ruby values onto these Go types.
package hocon
