package db2go2types

import (
	"errors"
	"fmt"
	"go/token"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Config configures [Generate].
type Config struct {
	// DSN is the PostgreSQL connection string: "postgres://user:pass@host:5432/db".
	// See [DSNFromEnv].
	DSN string
	// Schema to generate code for, e.g. "public". Required.
	Schema string
	// OutputDir receives models.go. Default "pkg/models".
	OutputDir string
	// Package is the name of the generated package. Default: the last element of OutputDir.
	Package string
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
const DefaultOutputDir = "pkg/models"

const defaultPackage = "models"

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
	if err := checkPackage(c.Package); err != nil {
		return c, err
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return c, nil
}

// packageName derives a package name from a directory, following Go style
// (lower case letters and digits only): "pkg/db-models" → "dbmodels".
func packageName(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	name := strings.ToLower(words(filepath.Base(abs)))
	if checkPackage(name) != nil {
		return defaultPackage
	}
	return name
}

// checkPackage reports whether name can be a package clause. The name is
// written into the generated source, so anything else is rejected.
func checkPackage(name string) error {
	if !token.IsIdentifier(name) || name == "_" || name == "main" {
		return fmt.Errorf("db2go2types: %q is not a valid package name", name)
	}
	return nil
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
