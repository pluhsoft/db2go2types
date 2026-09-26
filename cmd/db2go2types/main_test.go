package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pluhsoft/db2go2types/internal/pgtest"
)

func TestLoadEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	env := "# comment\n\nexport DB2GO_A=1\nDB2GO_B = \"two words\"\nDB2GO_C='x'\nDB2GO_SET=file\n"
	if err := os.WriteFile(path, []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"DB2GO_A", "DB2GO_B", "DB2GO_C"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
	t.Setenv("DB2GO_SET", "env")
	if err := loadEnv(path); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"DB2GO_A": "1", "DB2GO_B": "two words", "DB2GO_C": "x", "DB2GO_SET": "env"} {
		if got := os.Getenv(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	if err := loadEnv(filepath.Join(t.TempDir(), "missing")); err != nil {
		t.Errorf("missing file: %v", err)
	}
	if err := os.WriteFile(path, []byte("broken\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadEnv(path); err == nil || !strings.Contains(err.Error(), ":1:") {
		t.Errorf("broken line: %v", err)
	}
}

func TestRunFlags(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), nil, &out); err == nil || !strings.Contains(err.Error(), "-schema") {
		t.Errorf("missing -schema: %v", err)
	}
	if err := run(context.Background(), []string{"-schema", "x", "-enum-name", "bad"}, &out); err == nil {
		t.Error("bad -enum-name accepted")
	}
	out.Reset()
	if err := run(context.Background(), []string{"-version"}, &out); err != nil || !strings.HasPrefix(out.String(), "db2go2types ") {
		t.Errorf("-version: %q, %v", out.String(), err)
	}
}

func TestRun(t *testing.T) {
	db := pgtest.New(t, "../../examples/blog/schema.sql")
	dir := t.TempDir()
	var log bytes.Buffer
	err := run(context.Background(), []string{
		"-dsn", db.Config().ConnString(), "-schema", "blog", "-out", filepath.Join(dir, "models"),
		"-diagram", filepath.Join(dir, "db.md"), "-enum-name", "#A6D2FF=Blue", "-env", "", "-v",
	}, &log)
	if err != nil {
		t.Fatal(err)
	}
	models, err := os.ReadFile(filepath.Join(dir, "models", "models.go"))
	if err != nil || !bytes.Contains(models, []byte("package models")) || !bytes.Contains(models, []byte("LabelColorBlue")) {
		t.Errorf("models.go: %v\n%s", err, models)
	}
	if _, err := os.Stat(filepath.Join(dir, "db.md")); err != nil {
		t.Error(err)
	}
	if !strings.Contains(log.String(), "schema read") {
		t.Errorf("-v log: %s", log.String())
	}
}
