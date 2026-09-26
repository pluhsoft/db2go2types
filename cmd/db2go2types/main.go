// Command db2go2types generates Go types and repositories from a PostgreSQL schema.
//
//	db2go2types -schema public -out pkg/repository -diagram docs/db.md
//
// The connection string is taken from -dsn, DATABASE_URL or the POSTGRES_HOST,
// POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DATABASE and
// POSTGRES_SSLMODE variables. Variables from the -env file (default .env) are
// loaded first and never override the environment.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/pluhsoft/db2go2types"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stderr); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, "db2go2types:", err)
		}
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("db2go2types", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var (
		cfg     db2go2types.Config
		envFile string
		verbose bool
		version bool
	)
	cfg.EnumNames = map[string]string{}
	flags.StringVar(&cfg.DSN, "dsn", "", "PostgreSQL connection string (default: from the environment)")
	flags.StringVar(&cfg.Schema, "schema", "", "schema to generate code for (required)")
	flags.StringVar(&cfg.OutputDir, "out", db2go2types.DefaultOutputDir, "output directory")
	flags.StringVar(&cfg.Package, "package", "", "package name (default: last element of -out)")
	flags.StringVar(&cfg.PrimaryKey, "pk", db2go2types.DefaultPrimaryKey, "primary key column used by Get, skipped by Add and Update")
	flags.StringVar(&cfg.DiagramPath, "diagram", "", "write a Mermaid diagram of the schema to this Markdown file")
	flags.Func("enum-name", "Go name of an enum value, `value=Name`; repeatable, e.g. -enum-name '#A6D2FF=Blue'", func(s string) error {
		value, name, ok := strings.Cut(s, "=")
		if !ok || name == "" {
			return errors.New("want value=Name")
		}
		cfg.EnumNames[value] = name
		return nil
	})
	flags.StringVar(&envFile, "env", ".env", "file with environment variables, ignored if missing")
	flags.BoolVar(&verbose, "v", false, "log progress")
	flags.BoolVar(&version, "version", false, "print the version and exit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if version {
		fmt.Fprintln(stderr, "db2go2types", buildVersion())
		return nil
	}
	if cfg.Schema == "" {
		flags.Usage()
		return errors.New("-schema is required")
	}
	if err := loadEnv(envFile); err != nil {
		return err
	}
	if cfg.DSN == "" {
		dsn, err := db2go2types.DSNFromEnv()
		if err != nil {
			return err
		}
		cfg.DSN = dsn
	}
	if verbose {
		cfg.Logger = slog.New(slog.NewTextHandler(stderr, nil))
	}
	return db2go2types.Generate(ctx, cfg)
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

// loadEnv sets variables from a KEY=VALUE file that are not set yet.
// Blank lines, # comments, "export " prefixes and quoted values are supported.
func loadEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		if !ok {
			return fmt.Errorf("%s:%d: want KEY=VALUE", path, n)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		if _, set := os.LookupEnv(key); !set {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
