package db2go2types

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
)

// Generate connects to cfg.DSN, reads cfg.Schema and writes the generated
// package to cfg.OutputDir and the diagram to cfg.DiagramPath.
func Generate(ctx context.Context, cfg Config) error {
	if cfg.DSN == "" {
		return fmt.Errorf("db2go2types: Config.DSN is required")
	}
	conn, err := pgx.Connect(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("db2go2types: connect: %w", err)
	}
	defer conn.Close(context.WithoutCancel(ctx))
	return GenerateFrom(ctx, conn, cfg)
}

// GenerateFrom is [Generate] with an open connection; cfg.DSN is ignored.
func GenerateFrom(ctx context.Context, db Querier, cfg Config) error {
	cfg, err := cfg.withDefaults()
	if err != nil {
		return err
	}
	s, err := Inspect(ctx, db, cfg.Schema)
	if err != nil {
		return err
	}
	cfg.Logger.Info("schema read", "schema", s.Name, "tables", len(s.Tables), "enums", len(s.Enums))

	files, err := Render(s, cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("db2go2types: %w", err)
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(cfg.OutputDir, name)
		if err := os.WriteFile(path, files[name], 0o644); err != nil {
			return fmt.Errorf("db2go2types: %w", err)
		}
		cfg.Logger.Info("written", "file", path)
	}
	if cfg.DiagramPath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.DiagramPath), 0o755); err != nil {
			return fmt.Errorf("db2go2types: %w", err)
		}
		if err := os.WriteFile(cfg.DiagramPath, Diagram(s), 0o644); err != nil {
			return fmt.Errorf("db2go2types: %w", err)
		}
		cfg.Logger.Info("written", "file", cfg.DiagramPath)
	}
	return nil
}
