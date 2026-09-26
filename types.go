package db2go2types

// goTypes maps PostgreSQL types to Go types. Keys are data_type values from
// information_schema.columns and udt_name values used for array elements.
var goTypes = map[string]string{
	"smallint": "int16", "int2": "int16",
	"integer": "int", "int": "int", "int4": "int",
	"bigint": "int64", "int8": "int64",
	"serial": "int", "serial4": "int",
	"bigserial": "int64", "serial8": "int64",
	"boolean": "bool", "bool": "bool",
	"real": "float32", "float4": "float32",
	"double precision": "float64", "float8": "float64",
	// numeric loses the fraction; see the "Types" guide.
	"numeric": "int", "decimal": "int",
	"text": "string", "varchar": "string", "character varying": "string",
	"char": "string", "bpchar": "string", "character": "string",
	"uuid": "string", "citext": "string",
	"date":      "time.Time",
	"timestamp": "time.Time", "timestamp without time zone": "time.Time",
	"timestamptz": "time.Time", "timestamp with time zone": "time.Time",
	"time": "time.Time", "time without time zone": "time.Time",
	"timetz": "time.Time", "time with time zone": "time.Time",
	"bytea": "[]byte",
	"json":  "any", "jsonb": "any",
}

// goType returns the Go type of a column. Nullable columns are pointers,
// except nullable arrays, which are []any. Types without a mapping are any.
func (s *Schema) goType(c Column) string {
	base := s.baseGoType(c.ElementType())
	switch {
	case c.IsArray() && c.IsNullable:
		return "[]any"
	case c.IsArray():
		return "[]" + base
	case c.IsNullable:
		return "*" + base
	}
	return base
}

func (s *Schema) baseGoType(pgType string) string {
	if t, ok := goTypes[pgType]; ok {
		return t
	}
	if _, ok := s.enum(pgType); ok {
		return goName(pgType)
	}
	return "any"
}

// isEnumArray reports whether the column is an array of an enum of the schema.
// Such columns are selected as text[], which pgx scans into []EnumType.
func (s *Schema) isEnumArray(c Column) bool {
	if !c.IsArray() {
		return false
	}
	_, ok := s.enum(c.ElementType())
	return ok
}
