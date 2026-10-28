// Package db2go2types reads a PostgreSQL schema and generates Go types for it:
// a string type with constants per enum and a struct per table, ready to be
// scanned with pgx.
//
// Generate from Go, e.g. in a go:generate program:
//
//	err := db2go2types.Generate(ctx, db2go2types.Config{
//		DSN:       os.Getenv("DATABASE_URL"),
//		Schema:    "public",
//		OutputDir: "pkg/models",
//	})
//
// or with the command:
//
//	go run github.com/pluhsoft/db2go2types/cmd/db2go2types@latest -schema public -out pkg/models
//
// [Inspect], [Render] and [Diagram] are the steps of [Generate] for custom pipelines.
//
// Guides: https://pluhsoft.github.io/db2go2types/
package db2go2types
