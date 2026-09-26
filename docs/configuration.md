---
title: Configuration
nav_order: 2
---

# Configuration

| `Config` field | Flag          | Default          | Meaning                                                      |
| -------------- | ------------- | ---------------- | ------------------------------------------------------------ |
| `DSN`          | `-dsn`        | environment      | PostgreSQL connection string                                 |
| `Schema`       | `-schema`     | required         | schema to generate code for                                  |
| `OutputDir`    | `-out`        | `pkg/repository` | directory for `models.go` and `queries.go`                   |
| `Package`      | `-package`    | last part of `-out` | name of the generated package                             |
| `PrimaryKey`   | `-pk`         | `id`             | column that `Get` looks up and `Add`/`Update` skip           |
| `DiagramPath`  | `-diagram`    | none             | Markdown file with a Mermaid class diagram                    |
| `EnumNames`    | `-enum-name`  | none             | Go names for enum values: `-enum-name '#A6D2FF=Blue'`         |
| `Logger`       | `-v`          | silent           | progress log (`log/slog`)                                     |
| —              | `-env`        | `.env`           | file with environment variables, ignored if missing           |
| —              | `-version`    |                  | print the version                                             |

## Connection

Without `-dsn`, the command and [`DSNFromEnv`](https://pkg.go.dev/github.com/pluhsoft/db2go2types#DSNFromEnv) use:

1. `DATABASE_URL`, as is;
2. otherwise `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DATABASE`
   (all required) and `POSTGRES_SSLMODE` (optional). The password may contain any characters.

```sh
# .env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=app
POSTGRES_PASSWORD=secret
POSTGRES_DATABASE=app
```

Variables already set in the environment win over the file.

## What is read

Only the given schema: its base tables (views are skipped), their columns in table order, primary and
foreign keys, and the enum types defined in the schema. Enums from other schemas are mapped to `any`.

## Steps

`Generate` connects and calls three functions you can use on their own:

```go
s, err := db2go2types.Inspect(ctx, conn, "public")     // *Schema
files, err := db2go2types.Render(s, cfg)               // map[file name][]byte
md := db2go2types.Diagram(s)                           // Markdown with Mermaid
```

`GenerateFrom(ctx, conn, cfg)` is `Generate` with a connection you already have.
