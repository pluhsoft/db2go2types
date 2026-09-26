# Guide for AI agents

db2go2types reads a PostgreSQL schema and generates Go types for it. Repository generation is planned, not implemented.
The library is the root package; `cmd/db2go2types` is a thin command over it.

## Rules

- Dependencies: pgx only. Do not add others without an issue discussing it.
- Minimum Go version is in `go.mod`; do not use newer language or library features.
- Public API and the generated code follow SemVer: renaming a generated type or field breaks users.
- Every change: tests, `CHANGELOG.md` line under `## [Unreleased]`, docs in `docs/` if behavior changes.
- Generator changes: update the example with `go test . -run TestExample -update` and check that
  `examples/blog` tests still pass against PostgreSQL.
- Introspection SQL: `$n` parameters, never `fmt.Sprintf` with names. Names from the schema reach the generated
  source only through `goName`/`words`; values only as quoted strings (see docs/security.md).
- Work on an issue branch, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Map

| File                  | Contents                                                     |
| --------------------- | ------------------------------------------------------------ |
| `config.go`           | `Config`, defaults, `DSNFromEnv`                             |
| `generate.go`         | `Generate`, `GenerateFrom`: inspect, render, write           |
| `inspect.go`          | `Inspect`: tables, columns, keys, enums from the catalog     |
| `schema.go`           | `Schema`, `Table`, `Column`, keys, `Enum`                    |
| `types.go`            | PostgreSQL → Go type mapping                                 |
| `naming.go`           | Go names for tables, columns and enum values                 |
| `render.go`           | `Render`, `Diagram`, name collision checks                   |
| `templates/`          | `models.tmpl`, `diagram.tmpl`                                |
| `cmd/db2go2types`     | command, `.env` loading                                      |
| `examples/blog`       | schema, generated golden package, scan tests of the types     |
| `internal/pgtest`     | fresh PostgreSQL database per test                           |
| `internal/release`    | release tool used by CI                                      |

## Commands

```sh
export DB2GO_TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
go test . -run TestExample -update
```
