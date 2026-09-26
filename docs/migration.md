---
title: Migration
nav_order: 6
---

# Migration from sqlgenerator

db2go2types grew out of the `sqlgenerator` package with `GenerateRepository(customGeneratedPath)`.
The generated repositories keep their method names and signatures, so application code keeps compiling.

## Calling the generator

```go
// Before
sqlgenerator.GenerateRepository("")

// After
dsn, err := db2go2types.DSNFromEnv() // same POSTGRES_* variables
if err != nil {
	log.Fatal(err)
}
err = db2go2types.Generate(ctx, db2go2types.Config{
	DSN:         dsn,
	Schema:      "contentium",       // was hard-coded
	OutputDir:   "pkg/repository",   // was the argument
	DiagramPath: "db_diagram.md",    // was always written
	EnumNames: map[string]string{    // was hard-coded
		"#A6D2FF": "Blue", "#F87659": "Red", "#BCF1A5": "Green", "#FFD57A": "Yellow",
		"#90A5F9": "Purple", "#F9ABED": "Pink", "#70DFF9": "Turquoise", "#69E9C1": "Emerald",
	},
})
```

Or with the command:

```sh
db2go2types -schema contentium -out pkg/repository -diagram db_diagram.md \
  -enum-name '#A6D2FF=Blue' -enum-name '#F87659=Red' # …
```

`.env` is read by the command only; in your own program load it before calling `DSNFromEnv`.

## Differences in the generated code

- Enum constants follow the declaration order of the enum instead of alphabetical order.
- SQL identifiers are quoted: `"contentium"."posts"`, `"id"`.
- Arrays of enums are written as `$n::text[]::"schema"."enum"[]`; before, writing them failed.
- Columns of unknown types are `any` instead of code that did not compile.
- `Select…` returns `nil, err` on errors instead of an empty slice.
- Only columns, keys and enums of the given schema are read; views are skipped.
