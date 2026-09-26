package db2go2types

import "strings"

// Schema is a PostgreSQL schema read by [Inspect].
type Schema struct {
	// Name is the schema name, e.g. "public".
	Name string
	// Tables are sorted by name.
	Tables []Table
	// Enums are the enum types of the schema, sorted by name.
	Enums []Enum
}

// Table is a base table. Views are not included.
type Table struct {
	Schema      string
	Name        string
	Columns     []Column // in ordinal_position order
	PrimaryKeys []PrimaryKey
	ForeignKeys []ForeignKey
}

// Column is a table column.
type Column struct {
	Name string
	// Type is the PostgreSQL type: data_type from information_schema
	// ("integer", "timestamp with time zone"), the type name of enums and other
	// user-defined types ("post_status"), or "ARRAY:<element udt_name>" for
	// arrays ("ARRAY:int4", "ARRAY:post_status").
	Type       string
	IsNullable bool
}

// PrimaryKey is one column of a table's primary key.
type PrimaryKey struct {
	Name   string // constraint name
	Table  string
	Column string
}

// ForeignKey is one column of a foreign key constraint.
type ForeignKey struct {
	Name          string // constraint name
	Table         string
	Column        string
	ForeignSchema string
	ForeignTable  string
	ForeignColumn string
}

// Enum is an enum type.
type Enum struct {
	Name string
	// Values in declaration order.
	Values []string
}

// IsArray reports whether the column is an array.
func (c Column) IsArray() bool { return strings.HasPrefix(c.Type, arrayPrefix) }

// ElementType returns the element type of an array column, or Type otherwise.
func (c Column) ElementType() string {
	return strings.TrimPrefix(c.Type, arrayPrefix)
}

const arrayPrefix = "ARRAY:"

// enum returns the enum with the given PostgreSQL name.
func (s *Schema) enum(name string) (Enum, bool) {
	for _, e := range s.Enums {
		if e.Name == name {
			return e, true
		}
	}
	return Enum{}, false
}
