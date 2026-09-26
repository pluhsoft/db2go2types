package db2go2types

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
)

// Config configures [Generate].
type Config struct {
	// DSN is the PostgreSQL connection string: "postgres://user:pass@host:5432/db".
	// See [DSNFromEnv].
	DSN string
	// Schema to generate code for, e.g. "public". Required.
	Schema string
	// OutputDir receives models.go and queries.go. Default "pkg/repository".
	OutputDir string
	// Package is the name of the generated package. Default: the last element of OutputDir.
	Package string
	// PrimaryKey is the column that Get looks up and Add and Update skip.
	// Default "id".
	PrimaryKey string
	// DiagramPath is the file for the Mermaid class diagram of the schema.
	// Empty means no diagram.
	DiagramPath string
	// EnumNames gives Go names to enum values that make poor identifiers,
	// e.g. {"#A6D2FF": "Blue"} turns the constant ColorA6D2FF into ColorBlue.
	EnumNames map[string]string
	// Logger receives progress messages. Nil means no logging.
	Logger *slog.Logger
}

// Defaults of [Config].
const (
	DefaultOutputDir  = "pkg/repository"
	DefaultPrimaryKey = "id"
)

func (c Config) withDefaults() (Config, error) {
	if c.Schema == "" {
		return c, errors.New("db2go2types: Config.Schema is required")
	}
	if c.OutputDir == "" {
		c.OutputDir = DefaultOutputDir
	}
	if c.Package == "" {
		c.Package = packageName(c.OutputDir)
	}
	if c.PrimaryKey == "" {
		c.PrimaryKey = DefaultPrimaryKey
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return c, nil
}

// packageName derives a package name from a directory: "pkg/repository" → "repository".
func packageName(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	name := lowerFirst(words(filepath.Base(abs)))
	if name == "" {
		return "repository"
	}
	return name
}

// DSNFromEnv builds a connection string from environment variables.
// DATABASE_URL is used as is when set. Otherwise POSTGRES_HOST, POSTGRES_PORT,
// POSTGRES_USER, POSTGRES_PASSWORD and POSTGRES_DATABASE are required, and
// POSTGRES_SSLMODE is optional.
func DSNFromEnv() (string, error) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn, nil
	}
	vars := map[string]string{}
	var missing []string
	for _, key := range []string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DATABASE"} {
		vars[key] = os.Getenv(key)
		if vars[key] == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("db2go2types: set DATABASE_URL or the environment variables %v", missing)
	}
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(vars["POSTGRES_USER"], vars["POSTGRES_PASSWORD"]),
		Host:   net.JoinHostPort(vars["POSTGRES_HOST"], vars["POSTGRES_PORT"]),
		Path:   "/" + vars["POSTGRES_DATABASE"],
	}
	if mode := os.Getenv("POSTGRES_SSLMODE"); mode != "" {
		u.RawQuery = url.Values{"sslmode": {mode}}.Encode()
	}
	return u.String(), nil
}
