# Changelog

All notable changes are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow
[Semantic Versioning](https://semver.org).

## [Unreleased]

### Added

- `Generate`, `GenerateFrom`, `Inspect`, `Render`, `Diagram` and `DSNFromEnv`.
- `db2go2types` command with `-schema`, `-out`, `-package`, `-pk`, `-diagram`, `-enum-name` and `.env` support.
- Enum types, table structs, `Update…Params` and a repository per table with `Add`, `Update`, `Select`,
  `Get`, `Delete`, `Count`, `ExecuteQuery` and `ExecuteQueryRow` methods.
- Mermaid class diagram of the schema.
- `examples/blog`.

### Fixed (compared to the original `sqlgenerator` package)

- The schema is a setting instead of a hard-coded name; columns, keys and enums are read only from that schema.
- Introspection queries use parameters instead of string formatting.
- Views are skipped; composite and multi-column foreign keys are read correctly.
- Identifiers are quoted, so tables and columns like `user` or `order` work.
- `Add` no longer produces invalid Go when the primary key is not the first column.
- Arrays of enums can be written: parameters are sent as `text[]` and cast to the enum array.
- `timestamp with time zone`, `time with/without time zone` and `citext` are mapped; unknown types become `any`
  instead of code that does not compile.
- `time` and `errors` are imported only when used; output is `gofmt`-formatted and marked as generated.
- Enum values that are not identifiers (`#A6D2FF`, `in-progress`) produce valid constants; `EnumNames` replaces
  the hard-coded color names.
- Errors are returned instead of `panic` and `os.Exit`; passwords with special characters work in the DSN.
