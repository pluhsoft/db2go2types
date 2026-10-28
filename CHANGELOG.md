# Changelog

All notable changes are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow
[Semantic Versioning](https://semver.org).

## [Unreleased]

### Added

- `Generate`, `GenerateFrom`, `Inspect`, `Render`, `Diagram` and `DSNFromEnv`.
- `db2go2types` command with `-schema`, `-out`, `-package`, `-diagram`, `-enum-name` and `.env` support.
- Enum types and table structs in `models.go`; fields in column order for `pgx.RowToStructByPos`.
- Mermaid class diagram of the schema.
- Errors for names that collide in Go and for invalid package names; a hand-written `models.go` is never overwritten.
- `examples/blog`.

### Fixed (compared to the original `sqlgenerator` package)

- Repository generation (`queries.go`) is not part of this release; it is tracked in #1.

- The schema is a setting instead of a hard-coded name; columns, keys and enums are read only from that schema.
- Introspection queries use parameters instead of string formatting.
- Views are skipped; composite and multi-column foreign keys are read correctly.
- `timestamp with time zone`, `time with/without time zone` and `citext` are mapped; unknown types become `any`
  instead of code that does not compile.
- `time` is imported only when used; output is `gofmt`-formatted and marked as generated.
- Enum values that are not identifiers (`#A6D2FF`, `in-progress`) produce valid constants; `EnumNames` replaces
  the hard-coded color names.
- Errors are returned instead of `panic` and `os.Exit`; passwords with special characters work in the DSN.
