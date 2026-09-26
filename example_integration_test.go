package db2go2types_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pluhsoft/db2go2types"
	"github.com/pluhsoft/db2go2types/internal/pgtest"
)

var update = flag.Bool("update", false, "rewrite golden files")

// exampleConfig generates examples/blog; the example's tests run the generated code.
func exampleConfig(dir string) db2go2types.Config {
	return db2go2types.Config{
		Schema:      "blog",
		OutputDir:   filepath.Join(dir, "models"),
		DiagramPath: filepath.Join(dir, "schema.md"),
		EnumNames:   map[string]string{"#A6D2FF": "Blue", "#F87659": "Red", "#BCF1A5": "Green"},
	}
}

// TestExample checks that examples/blog matches the generator output for examples/blog/schema.sql.
func TestExample(t *testing.T) {
	db := pgtest.New(t, "examples/blog/schema.sql")
	dir := "examples/blog"
	if !*update {
		dir = t.TempDir()
	}
	cfg := exampleConfig(dir)
	cfg.DSN = db.Config().ConnString()
	if err := db2go2types.Generate(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	if *update {
		return
	}
	for _, file := range []string{"models/models.go", "schema.md"} {
		got, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.ReadFile(filepath.Join("examples/blog", file))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("examples/blog/%s is out of date; run: go test . -run TestExample -update\n%s", file, got)
		}
	}
}

func TestInspect(t *testing.T) {
	db := pgtest.New(t, "examples/blog/schema.sql")
	s, err := db2go2types.Inspect(context.Background(), db, "blog")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, table := range s.Tables {
		names = append(names, table.Name)
	}
	if got := fmt.Sprint(names); got != "[authors post_likes posts]" {
		t.Errorf("tables = %s; views must be skipped", got)
	}
	likes := s.Tables[1]
	if len(likes.PrimaryKeys) != 2 || len(likes.ForeignKeys) != 2 || likes.ForeignKeys[1].ForeignTable != "posts" {
		t.Errorf("post_likes keys = %+v %+v", likes.PrimaryKeys, likes.ForeignKeys)
	}
	if got := fmt.Sprint(s.Enums); got != "[{label_color [#A6D2FF #F87659 #BCF1A5]} {post_status [draft published archived]}]" {
		t.Errorf("enums = %s", got)
	}
	if _, err := db2go2types.Inspect(context.Background(), db, "missing"); err == nil {
		t.Error("missing schema accepted")
	}
}
