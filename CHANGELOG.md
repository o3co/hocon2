# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

## [0.8.0] - 2026-08-26

### Changed — go.hocon 1.13.0: the four-implementation spec-correction release

The parser and adapters move from v1.12.0 to v1.13.0, the lockstep release that
aligns all four sibling implementations with Lightbend on a group of spec
corrections. What reaches hocon2:

- **Triple-quoted strings preserve CR/CRLF verbatim** (previously normalized to
  LF). A conversion whose HOCON input carries `\r\n` inside `"""…"""` now
  reproduces it in the output — BREAKING if a downstream consumer relied on the
  old normalization.
- **Self-reference and delayed-merge corrections (S9.2 / S13.12 / S13a.12)** —
  edge cases around `a = ${a}` chains and object-over-array overrides now
  resolve the way Lightbend does, so some inputs produce different
  (now-correct) trees, and some previously-accepted cycles are now errors.
- The `*2hocon` reverse converters now quote a bare `include` key
  (`"include" = …`) in their output, which Lightbend would otherwise reject on
  re-parse.

Full detail in the
[go.hocon 1.13.0 release notes](https://github.com/o3co/go.hocon/releases/tag/v1.13.0).

### Changed — go.hocon 1.12.0, which changes what `.env` input is accepted

The parser and adapters move from v1.11.0 to v1.12.0. Most of that release is
adapter hardening that reaches hocon2 without any behaviour a user would notice
— a silent U+FFFD substitution stopped, a depth limit added, a YAML stream whose
first document was empty no longer discarding the file.

One change is **BREAKING in both directions** and does reach `env2hocon`:

- the prefix filter now runs **before** validation, so a `.env` that used to be
  rejected may now load
- a variable name containing whitespace or `#` is now **refused**, so a `.env`
  that used to load may now fail — quote or fix the name

Whitespace there is the Unicode `White_Space` property, pinned across the four
implementations rather than inherited from each language's stdlib
([xx.hocon#78](https://github.com/o3co/xx.hocon/issues/78)). Full detail in the
[go.hocon 1.12.0 release notes](https://github.com/o3co/go.hocon/releases/tag/v1.12.0).

## [0.7.0] - 2026-07-26

### Added — reverse conversion (foreign format → HOCON)

- **Four new commands read a JSON / YAML / TOML / Properties file and render
  HOCON**: `json2hocon`, `yaml2hocon`, `toml2hocon`, `properties2hocon`. This
  completes the bidirectional promise (`hocon2X` was one-way); an existing config
  in another format can now be imported into HOCON.
- Input parsing reuses the [go.hocon adapters](https://github.com/o3co/go.hocon/tree/main/adapters)
  so the format-mapping rules (TOML dates → strings, YAML 1.2 core schema, BOM
  stripping) are shared with the parser libraries rather than a second copy that
  could drift. Output uses the new `Config.RenderHOCON()` emitter
  (go.hocon v1.11.0).
- The conversion preserves types across the round trip: a Properties value or a
  JSON string `"8080"` is quoted so it re-parses as a string, not a number. A
  `${...}` in an input value is emitted **literally** — foreign data is data, not
  a substitution to resolve.
- The reverse commands mirror the forward ones: `-o` / `-overwrite` / `-validate`,
  stdin or one FILE, and the same atomic output write. There is no `-env-file`
  (nothing to resolve). Reverse takes a single input, not a merge set.
- Note: YAML scalar resolution is the YAML library's answer, not a guarantee here
  — whether a bare `010` is `8` or `10` depends on the parser. Quote values that
  must survive verbatim.

### Changed

- Bump `go.hocon` from v1.10.0 to **v1.11.0** (adds the `RenderHOCON()` emitter
  and adapter fixes) and add `go.hocon/adapters` **v1.11.0** as a direct
  dependency (the reverse decoders reuse it for input parsing).

## [0.6.0] - 2026-07-25

### Changed

- Bump `go.hocon` parser from v1.1.0 to **v1.10.0** — brings Lightbend-parity
  fixes for include-scope substitution resolution, `+=` accumulation across
  includes, self-referential append chains, key-path whitespace/trailing-dot
  handling, and improved `Unmarshal` (null preservation in `map[string]any`,
  numeric-keyed objects to slices). v1.9.0–v1.10.0 add the S3.1/S3.5/S19.8/S11.7/S8.1
  spec corrections and the full `java.util.Properties` include syntax (S23.5/S23.6).
  The v1.10.0 `adapters/` module is a separate Go module and is not pulled in.

### Fixed

- `-validate` now stops after parsing, making it a pure syntax check as documented (#16)
- Failed encodes no longer corrupt the output file — output is written to a temp file and renamed into place atomically (#15)
- `-env-file` now accepts `export KEY=value` lines as written by direnv/dotenv (#14)

## [0.5.0] - 2026-04-07

### Changed

- Bump `go.hocon` parser from v0.3.0 to v1.1.0 — brings scalar raw-string representation, improved substitution handling, and full Lightbend conformance

## [0.3.0] - 2026-03-20

### Added

- `-compact` and `-indent` options for JSON output formatting
- `-o` output file option with `-overwrite` safety flag
- Windows binary releases
- This CHANGELOG

## [0.2.0] - 2026-03-20

### Changed

- Module path changed from `go.hocon2` to `hocon2`

### Added

- Japanese README

## [0.1.0] - 2026-03-20

### Added

- HOCON to JSON, YAML, TOML, and Properties conversion
- Multi-file merge with right-precedence
- Stdin and file input support
- Lightbend conformance tests (equiv01–equiv05)
- GoReleaser configuration for Linux and macOS
