# Security policy

## Supported versions

The latest minor release receives security fixes.

## Reporting a vulnerability

Do not open a public issue. Report privately via
[GitHub security advisories](https://github.com/pluhsoft/db2go2types/security/advisories/new).
You get an answer within 7 days.

## Scope

The generator connects to a database and writes Go source. In scope: code injection into generated
files through schema names, leaking credentials, overwriting files outside the output. See the
[security guide](https://pluhsoft.github.io/db2go2types/security) for safe usage.
