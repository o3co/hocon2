# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased]

### Changed

- Bump `go.hocon` parser from v1.1.0 to v1.8.0 — brings Lightbend-parity fixes for include-scope substitution resolution, `+=` accumulation across includes, self-referential append chains, key-path whitespace/trailing-dot handling, and improved `Unmarshal` (null preservation in `map[string]any`, numeric-keyed objects to slices)

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
