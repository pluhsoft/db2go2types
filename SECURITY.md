# Security policy

## Supported versions

The latest minor release receives security fixes.

## Reporting a vulnerability

Do not open a public issue. Report privately via
[GitHub security advisories](https://github.com/pluhsoft/db2go2types/security/advisories/new).
You get an answer within 7 days.

## Generated code

Repository methods that take a `where` clause or a custom query (`Select…`, `Update…`, `Delete…`,
`Count…`, `Get…` and `Add…` with `customQuery`) put that string into the SQL as is.
Passing user input there is an SQL injection in your application, not a vulnerability of
db2go2types. Use `ExecuteQuery…` with `$1` parameters instead. See the
[security guide](https://pluhsoft.github.io/db2go2types/security).
