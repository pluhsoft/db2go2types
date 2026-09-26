# db2go2types

[![Go Reference](https://pkg.go.dev/badge/github.com/pluhsoft/db2go2types.svg)](https://pkg.go.dev/github.com/pluhsoft/db2go2types)
[![CI](https://github.com/pluhsoft/db2go2types/actions/workflows/ci.yml/badge.svg)](https://github.com/pluhsoft/db2go2types/actions/workflows/ci.yml)

Generates Go types and a repository layer from a live PostgreSQL schema. The database is the source
of truth: change the schema, run the generator, and the compiler shows every place to update.

- A Go type and constants per enum, a struct per table, `Update…Params` for writes
- A repository per table on [pgx v5](https://github.com/jackc/pgx): `Add`, `Update`, `Select`, `Get`,
  `Delete`, `Count`, `ExecuteQuery`, `ExecuteQueryRow`
- Works with `*pgxpool.Pool`, `*pgx.Conn` and transactions (`pgx.Tx`)
- Enum arrays, nullable columns, reserved words as names
- Mermaid diagram of the schema
- Library API and a command for `go:generate`

## Install

```sh
go install github.com/pluhsoft/db2go2types/cmd/db2go2types@latest
```

Requires Go 1.25 or later. Tested with PostgreSQL 16 and 17.

## Example

```sql
CREATE TYPE blog.post_status AS ENUM ('draft', 'published', 'archived');
CREATE TABLE blog.posts (
    id        serial PRIMARY KEY,
    author_id integer NOT NULL REFERENCES blog.authors (id),
    title     text NOT NULL,
    status    blog.post_status NOT NULL DEFAULT 'draft',
    tags      text[]
);
```

```sh
DATABASE_URL=postgres://user:pass@localhost:5432/app db2go2types -schema blog -out pkg/repository
```

```go
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

type Posts struct {
	Id       int        `validate:"required"`
	AuthorId int        `validate:"required"`
	Title    string     `validate:"required"`
	Status   PostStatus `validate:"required"`
	Tags     []any
}
```

```go
posts := repository.NewPostsRepository(pool)

post, err := posts.AddPosts(ctx, repository.UpdatePostsParams{
	AuthorId: 1, Title: "Hello", Status: repository.PostStatusDraft,
}, nil)

published, err := posts.ExecuteQueryPosts(ctx,
	`SELECT id, author_id, title, status, tags FROM blog.posts WHERE status = $1`,
	repository.PostStatusPublished)
```

The complete generated package with tests: [`examples/blog`](examples/blog).

## Documentation

- [Guide](https://pluhsoft.github.io/db2go2types/): configuration, generated code, types, security
- [Migration](https://pluhsoft.github.io/db2go2types/migration) from the original `sqlgenerator` package
- [API reference](https://pkg.go.dev/github.com/pluhsoft/db2go2types)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE)
