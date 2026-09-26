---
title: Quick start
nav_order: 1
---

# db2go2types

Go types and a [pgx](https://github.com/jackc/pgx) repository layer generated from a live PostgreSQL
schema. The database is the source of truth; the generated package is never edited by hand.

## Command

```sh
go install github.com/pluhsoft/db2go2types/cmd/db2go2types@latest
db2go2types -schema public -out pkg/repository -diagram docs/db.md
```

The connection string comes from `-dsn`, `DATABASE_URL` or the `POSTGRES_*` variables; a `.env` file
in the current directory is read too. See [Configuration](configuration).

## go:generate

```go
//go:generate go run github.com/pluhsoft/db2go2types/cmd/db2go2types@v0.1.0 -schema public -out .
package repository
```

Pin the version so that everyone generates the same code.

## Library

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/pluhsoft/db2go2types"
)

func main() {
	err := db2go2types.Generate(context.Background(), db2go2types.Config{
		DSN:       os.Getenv("DATABASE_URL"),
		Schema:    "public",
		OutputDir: "pkg/repository",
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Using the generated code

```go
pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
posts := repository.NewPostsRepository(pool)

post, err := posts.GetPosts(ctx, 42, nil)
if errors.Is(err, pgx.ErrNoRows) {
	// not found
}
```

## Next

- [Configuration](configuration): options, flags, environment
- [Generated code](generated-code): models and repository methods, transactions
- [Types](types): PostgreSQL → Go mapping and its limits
- [Security](security): where clauses and SQL injection
- [Migration](migration) from the original `sqlgenerator` package
