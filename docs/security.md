---
title: Security
nav_order: 5
---

# Security

## Database access

- The generator only reads the system catalog (`pg_namespace`, `pg_type`, `pg_enum`, `pg_constraint`,
  `information_schema`) with parameterized queries. It never reads table data and never writes.
- Use a dedicated read-only role; it needs `CONNECT` on the database and `USAGE` on the schema:

  ```sql
  CREATE ROLE db2go LOGIN PASSWORD '…';
  GRANT CONNECT ON DATABASE app TO db2go;
  GRANT USAGE ON SCHEMA public TO db2go;
  ```

  Without `SELECT` on a table, `information_schema.columns` hides its columns, so grant `SELECT` or
  run against a development database with the same schema.
- Prefer a local or CI database created from migrations over production.
- Use `sslmode=verify-full` for remote servers (`POSTGRES_SSLMODE` or the DSN).
- Keep passwords out of the command line, where other users can see them in the process list:
  use `.env` or environment variables. Do not commit `.env`.

## Generated code

Names from the schema can be crafted by whoever can change it. The generator treats them as data:

- table, column and enum names become identifiers only after dropping every character that cannot
  be in a Go identifier;
- enum values are written as quoted Go strings;
- the package name must be a valid Go identifier;
- no comments or other text from the database is written.

So a schema cannot inject Go code into `models.go`. Review the generated diff in pull requests anyway.

## Files

`Generate` refuses to overwrite a `models.go` that it did not generate.

Report vulnerabilities privately, see [SECURITY.md](https://github.com/pluhsoft/db2go2types/blob/main/SECURITY.md).
