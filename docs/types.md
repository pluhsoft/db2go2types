---
title: Types
nav_order: 4
---

# Types

| PostgreSQL                                            | Go            |
| ----------------------------------------------------- | ------------- |
| `smallint`                                            | `int16`       |
| `integer`, `serial`                                   | `int`         |
| `bigint`, `bigserial`                                 | `int64`       |
| `real`                                                | `float32`     |
| `double precision`                                    | `float64`     |
| `numeric`, `decimal`                                  | `int` (see below) |
| `boolean`                                             | `bool`        |
| `text`, `varchar`, `char`, `uuid`, `citext`           | `string`      |
| `date`, `timestamp`, `timestamptz`, `time`, `timetz`  | `time.Time`   |
| `bytea`                                               | `[]byte`      |
| `json`, `jsonb`                                       | `any`         |
| enum of the schema                                    | its Go type   |
| array of T                                            | `[]T`         |
| anything else                                         | `any`         |

Nullable columns become pointers (`*string`), nullable arrays become `[]any`.

## Known limits

These keep the output compatible with the original generator and are open questions for v1:

- `numeric` is `int`: values with a fraction fail to scan.
- Nullable arrays are `[]any` instead of `[]T`.
- Enums from other schemas and domains are `any`.
