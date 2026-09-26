# Guide for AI agents

db2go2types reads a PostgreSQL schema and generates Go types and pgx repositories for it.
The library is the root package; `cmd/db2go2types` is a thin command over it.

## Rules

- Dependencies: pgx only. Do not add others without an issue discussing it.
- Minimum Go version is in `go.mod`; do not use newer language or library features.
- Public API and the generated code follow SemVer: a change in generated method signatures breaks users.
- Every change: tests, `CHANGELOG.md` line under `## [Unreleased]`, docs in `docs/` if behavior changes.
- Generator changes: update the example with `go test . -run TestExample -update` and check that
  `examples/blog` tests still pass against PostgreSQL.
- SQL in the generator and introspection: quoted identifiers, `$n` parameters, never `fmt.Sprintf` with names.
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
| `render.go`           | `Render`, `Diagram`, SQL building                            |
| `templates/`          | `models.tmpl`, `queries.tmpl`, `diagram.tmpl`                |
| `cmd/db2go2types`     | command, `.env` loading                                      |
| `examples/blog`       | schema, generated golden package, tests of the generated code |
| `internal/pgtest`     | fresh PostgreSQL database per test                           |
| `internal/release`    | release tool used by CI                                      |

## Commands

```sh
export DB2GO_TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
go test -race ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
go test . -run TestExample -update
```
