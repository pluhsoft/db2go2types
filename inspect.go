package db2go2types

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Querier runs a query. *pgx.Conn, *pgxpool.Pool and pgx.Tx implement it.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Inspect reads the tables, columns, keys and enum types of a PostgreSQL schema.
func Inspect(ctx context.Context, db Querier, schema string) (*Schema, error) {
	var exists bool
	if err := queryEach(ctx, db, `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)`,
		[]any{schema}, func(r pgx.Rows) error { return r.Scan(&exists) }); err != nil {
		return nil, fmt.Errorf("db2go2types: check schema: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("db2go2types: schema %q does not exist", schema)
	}

	s := &Schema{Name: schema}
	steps := []struct {
		what string
		read func(context.Context, Querier, *Schema) error
	}{
		{"enums", readEnums},
		{"tables", readTables},
		{"columns", readColumns},
		{"primary keys", readPrimaryKeys},
		{"foreign keys", readForeignKeys},
	}
	for _, step := range steps {
		if err := step.read(ctx, db, s); err != nil {
			return nil, fmt.Errorf("db2go2types: read %s of schema %q: %w", step.what, schema, err)
		}
	}
	return s, nil
}

func queryEach(ctx context.Context, db Querier, sql string, args []any, scan func(pgx.Rows) error) error {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func readEnums(ctx context.Context, db Querier, s *Schema) error {
	return queryEach(ctx, db, `
		SELECT t.typname, e.enumlabel
		FROM pg_type t
		JOIN pg_enum e ON e.enumtypid = t.oid
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = $1
		ORDER BY t.typname, e.enumsortorder`,
		[]any{s.Name}, func(r pgx.Rows) error {
			var name, value string
			if err := r.Scan(&name, &value); err != nil {
				return err
			}
			if n := len(s.Enums); n == 0 || s.Enums[n-1].Name != name {
				s.Enums = append(s.Enums, Enum{Name: name})
			}
			last := &s.Enums[len(s.Enums)-1]
			last.Values = append(last.Values, value)
			return nil
		})
}

func readTables(ctx context.Context, db Querier, s *Schema) error {
	return queryEach(ctx, db, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name`,
		[]any{s.Name}, func(r pgx.Rows) error {
			var name string
			if err := r.Scan(&name); err != nil {
				return err
			}
			s.Tables = append(s.Tables, Table{Schema: s.Name, Name: name})
			return nil
		})
}

func (s *Schema) table(name string) *Table {
	for i := range s.Tables {
		if s.Tables[i].Name == name {
			return &s.Tables[i]
		}
	}
	return nil
}

func readColumns(ctx context.Context, db Querier, s *Schema) error {
	return queryEach(ctx, db, `
		SELECT table_name, column_name, data_type, is_nullable, udt_name
		FROM information_schema.columns
		WHERE table_schema = $1
		ORDER BY table_name, ordinal_position`,
		[]any{s.Name}, func(r pgx.Rows) error {
			var table, name, dataType, nullable, udtName string
			if err := r.Scan(&table, &name, &dataType, &nullable, &udtName); err != nil {
				return err
			}
			t := s.table(table)
			if t == nil { // a view
				return nil
			}
			switch dataType {
			case "ARRAY":
				// udt_name of an array is the element type with a leading underscore: _int4.
				dataType = arrayPrefix + strings.TrimPrefix(udtName, "_")
			case "USER-DEFINED":
				dataType = udtName
			}
			t.Columns = append(t.Columns, Column{Name: name, Type: dataType, IsNullable: nullable == "YES"})
			return nil
		})
}

func readPrimaryKeys(ctx context.Context, db Querier, s *Schema) error {
	return queryEach(ctx, db, `
		SELECT c.conname, t.relname, a.attname
		FROM pg_constraint c
		JOIN pg_class t ON t.oid = c.conrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		CROSS JOIN LATERAL unnest(c.conkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = k.attnum
		WHERE c.contype = 'p' AND n.nspname = $1
		ORDER BY t.relname, k.ord`,
		[]any{s.Name}, func(r pgx.Rows) error {
			var pk PrimaryKey
			if err := r.Scan(&pk.Name, &pk.Table, &pk.Column); err != nil {
				return err
			}
			if t := s.table(pk.Table); t != nil {
				t.PrimaryKeys = append(t.PrimaryKeys, pk)
			}
			return nil
		})
}

func readForeignKeys(ctx context.Context, db Querier, s *Schema) error {
	return queryEach(ctx, db, `
		SELECT c.conname, src.relname, sa.attname, dn.nspname, dst.relname, da.attname
		FROM pg_constraint c
		JOIN pg_class src ON src.oid = c.conrelid
		JOIN pg_namespace n ON n.oid = src.relnamespace
		JOIN pg_class dst ON dst.oid = c.confrelid
		JOIN pg_namespace dn ON dn.oid = dst.relnamespace
		CROSS JOIN LATERAL unnest(c.conkey, c.confkey) WITH ORDINALITY AS k(src_att, dst_att, ord)
		JOIN pg_attribute sa ON sa.attrelid = c.conrelid AND sa.attnum = k.src_att
		JOIN pg_attribute da ON da.attrelid = c.confrelid AND da.attnum = k.dst_att
		WHERE c.contype = 'f' AND n.nspname = $1
		ORDER BY src.relname, c.conname, k.ord`,
		[]any{s.Name}, func(r pgx.Rows) error {
			var fk ForeignKey
			if err := r.Scan(&fk.Name, &fk.Table, &fk.Column, &fk.ForeignSchema, &fk.ForeignTable, &fk.ForeignColumn); err != nil {
				return err
			}
			if t := s.table(fk.Table); t != nil {
				t.ForeignKeys = append(t.ForeignKeys, fk)
			}
			return nil
		})
}
