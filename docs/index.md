---
title: Quick start
nav_order: 1
---

# db2go2types

Go types generated from a live PostgreSQL schema. The database is the source of truth; the generated
package is never edited by hand.

## Command

```sh
go install github.com/pluhsoft/db2go2types/cmd/db2go2types@latest
db2go2types -schema public -out pkg/models -diagram docs/db.md
```

The connection string comes from `-dsn`, `DATABASE_URL` or the `POSTGRES_*` variables; a `.env` file
in the current directory is read too. See [Configuration](configuration).

## go:generate

```go
//go:generate go run github.com/pluhsoft/db2go2types/cmd/db2go2types@v0.1.0 -schema public -out .
package models
```

Pin the version so that everyone generates the same code.

## Library

```go
err := db2go2types.Generate(context.Background(), db2go2types.Config{
	DSN:       os.Getenv("DATABASE_URL"),
	Schema:    "public",
	OutputDir: "pkg/models",
})
```

## Using the types

Fields follow the column order of the table, so pgx can scan rows by position:

```go
rows, err := conn.Query(ctx, `SELECT * FROM public.posts WHERE author_id = $1`, authorID)
posts, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Posts])
```

## Next

- [Configuration](configuration): options, flags, environment
- [Generated types](generated-types): enums, structs, names
- [Types](types): PostgreSQL → Go mapping and its limits
- [Security](security): database access, generated files
- [Migration](migration) from the original `sqlgenerator` package
