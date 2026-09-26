---
title: Migration
nav_order: 6
---

# Migration from sqlgenerator

db2go2types grew out of the `sqlgenerator` package with `GenerateRepository(customGeneratedPath)`.
This version generates the types (`models.go`) only; generating the repository layer (`queries.go`)
is planned in [#1](https://github.com/pluhsoft/db2go2types/issues/1). Until then keep the old generator for `queries.go` or write the queries by hand.

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

## Differences in models.go

- `Update…Params` structs are not generated; they belong to the repository layer.
- Enum constants follow the declaration order of the enum instead of alphabetical order.
- Columns of unknown types are `any` instead of code that did not compile.
- `time` is imported only when used.
- Only columns and enums of the given schema are read; views are skipped.
