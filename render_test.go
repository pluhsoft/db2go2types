package db2go2types

import (
	"strings"
	"testing"
)

func TestGoName(t *testing.T) {
	for in, want := range map[string]string{
		"id_user": "IdUser", "posts": "Posts", "in-progress": "InProgress", "#A6D2FF": "A6D2FF",
		"2fa_codes": "X2faCodes", "": "X", "__": "X", "é_t": "ÉT",
	} {
		if got := goName(in); got != want {
			t.Errorf("goName(%q) = %q, want %q", in, got, want)
		}
	}
	if got := enumConstName("Color", "#A6D2FF", map[string]string{"#A6D2FF": "blue"}); got != "ColorBlue" {
		t.Errorf("custom enum name = %q", got)
	}
	if got := enumConstName("Mode", "", nil); got != "ModeEmpty" {
		t.Errorf("empty enum value = %q", got)
	}
	if got := packageName("pkg/db-models"); got != "dbModels" {
		t.Errorf("packageName = %q", got)
	}
}

func TestGoType(t *testing.T) {
	s := &Schema{Enums: []Enum{{Name: "mood", Values: []string{"ok"}}}}
	for _, tc := range []struct {
		col  Column
		want string
	}{
		{Column{Type: "integer"}, "int"},
		{Column{Type: "timestamp with time zone", IsNullable: true}, "*time.Time"},
		{Column{Type: "mood"}, "Mood"},
		{Column{Type: "ARRAY:mood"}, "[]Mood"},
		{Column{Type: "ARRAY:int8"}, "[]int64"},
		{Column{Type: "ARRAY:text", IsNullable: true}, "[]any"},
		{Column{Type: "inet"}, "any"},
		{Column{Type: "jsonb", IsNullable: true}, "*any"},
	} {
		if got := s.goType(tc.col); got != tc.want {
			t.Errorf("goType(%+v) = %s, want %s", tc.col, got, tc.want)
		}
	}
}

func TestRenderEdgeCases(t *testing.T) {
	s := &Schema{
		Name:  "app",
		Enums: []Enum{{Name: "kind", Values: []string{"a-b", "a_b", "a b"}}},
		Tables: []Table{
			{Schema: "app", Name: "counters", Columns: []Column{{Name: "id", Type: "integer"}}},
			{Schema: "app", Name: "user", Columns: []Column{
				{Name: "id", Type: "bigint"},
				{Name: "na`me", Type: "text"},
				{Name: "kinds", Type: "ARRAY:kind"},
			}},
		},
	}
	files, err := Render(s, Config{Package: "db"})
	if err != nil {
		t.Fatal(err)
	}
	models, queries := string(files[ModelsFile]), string(files[QueriesFile])
	for _, want := range []string{
		"package db\n",
		"KindAB  Kind = \"a-b\"", "KindAB2 Kind = \"a_b\"", "KindAB3 Kind = \"a b\"",
	} {
		if !strings.Contains(models, want) {
			t.Errorf("models.go has no %q:\n%s", want, models)
		}
	}
	if strings.Contains(models, `"time"`) {
		t.Error("models.go imports time without time columns")
	}
	for _, want := range []string{
		"`INSERT INTO \"app\".\"counters\" DEFAULT VALUES RETURNING \"id\"`",
		`errors.New("Counters has no columns to update")`,
		`"INSERT INTO \"app\".\"user\" (\"na` + "`" + `me\", \"kinds\") VALUES ($1, $2::text[]::\"app\".\"kind\"[])`,
		`\"kinds\"::text[]`,
	} {
		if !strings.Contains(queries, want) {
			t.Errorf("queries.go has no %s", want)
		}
	}

	empty, err := Render(&Schema{Name: "empty"}, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if q := string(empty[QueriesFile]); strings.Contains(q, `"errors"`) || !strings.Contains(q, "package repository") {
		t.Errorf("queries.go of an empty schema:\n%s", q)
	}
}

func TestDiagram(t *testing.T) {
	s := &Schema{Tables: []Table{{Name: "posts",
		Columns:     []Column{{Name: "tags", Type: "ARRAY:text", IsNullable: true}},
		ForeignKeys: []ForeignKey{{Table: "posts", Column: "author_id", ForeignTable: "authors"}},
	}}}
	got := string(Diagram(s))
	for _, want := range []string{"class posts {", "text[] tags (nullable)", "posts --> authors : author_id"} {
		if !strings.Contains(got, want) {
			t.Errorf("diagram has no %q:\n%s", want, got)
		}
	}
}

func TestConfig(t *testing.T) {
	if _, err := (Config{}).withDefaults(); err == nil {
		t.Error("empty schema accepted")
	}
	c, err := Config{Schema: "public"}.withDefaults()
	if err != nil || c.OutputDir != DefaultOutputDir || c.Package != "repository" || c.PrimaryKey != "id" || c.Logger == nil {
		t.Errorf("defaults = %+v, %v", c, err)
	}
}

func TestDSNFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	for _, key := range []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DATABASE", "POSTGRES_SSLMODE"} {
		t.Setenv(key, "")
	}
	if _, err := DSNFromEnv(); err == nil || !strings.Contains(err.Error(), "POSTGRES_HOST") {
		t.Errorf("missing variables: %v", err)
	}
	t.Setenv("POSTGRES_HOST", "db")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "app")
	t.Setenv("POSTGRES_PASSWORD", "p@ss/word")
	t.Setenv("POSTGRES_DATABASE", "shop")
	t.Setenv("POSTGRES_SSLMODE", "disable")
	if dsn, err := DSNFromEnv(); err != nil || dsn != "postgres://app:p%40ss%2Fword@db:5432/shop?sslmode=disable" {
		t.Errorf("DSN = %s, %v", dsn, err)
	}
	t.Setenv("DATABASE_URL", "postgres://x")
	if dsn, _ := DSNFromEnv(); dsn != "postgres://x" {
		t.Errorf("DATABASE_URL ignored: %s", dsn)
	}
}
