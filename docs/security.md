---
title: Security
nav_order: 5
---

# Security

## where clauses and custom queries

`Select…`, `Update…`, `Delete…`, `Count…` take a `where` string, and `Add…`, `Get…` take a
`customQuery`. The string becomes part of the SQL as is. Building it from user input is an
SQL injection:

```go
// Never:
posts.SelectPosts(ctx, ptr("WHERE title = '"+r.FormValue("title")+"'"))
```

Use `ExecuteQuery…` with parameters instead:

```go
posts.ExecuteQueryPosts(ctx,
	`SELECT id, author_id, title, status, tags FROM blog.posts WHERE title = $1`,
	r.FormValue("title"))
```

Constant strings are safe: `ptr("WHERE status = 'published' ORDER BY id")`.

## Generator

- Introspection only reads the catalog with parameterized queries. A read-only user is enough.
- Identifiers in generated SQL are quoted, so names from the schema cannot change the query.
- Do not commit `.env` files with passwords; `.gitignore` of this repository excludes `.env`.

Report vulnerabilities privately, see [SECURITY.md](https://github.com/pluhsoft/db2go2types/blob/main/SECURITY.md).
