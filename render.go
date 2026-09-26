package db2go2types

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"strconv"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"goString": goString,
	"pgType": func(c Column) string {
		if c.IsArray() {
			return c.ElementType() + "[]"
		}
		return c.Type
	},
}).ParseFS(templateFS, "templates/*.tmpl"))

// File names written to Config.OutputDir.
const (
	ModelsFile  = "models.go"
	QueriesFile = "queries.go"
)

// Render generates the Go files for a schema: [ModelsFile] with enum types,
// table structs and Update…Params, and [QueriesFile] with a repository per
// table. The result maps file names to gofmt-formatted sources.
// Only Package, PrimaryKey and EnumNames of cfg are used.
func Render(s *Schema, cfg Config) (map[string][]byte, error) {
	if cfg.Package == "" {
		cfg.Package = "repository"
	}
	if cfg.PrimaryKey == "" {
		cfg.PrimaryKey = DefaultPrimaryKey
	}
	data := newFileData(s, cfg)
	files := map[string][]byte{}
	for name, tmpl := range map[string]string{ModelsFile: "models.tmpl", QueriesFile: "queries.tmpl"} {
		var buf bytes.Buffer
		if err := templates.ExecuteTemplate(&buf, tmpl, data); err != nil {
			return nil, fmt.Errorf("db2go2types: render %s: %w", name, err)
		}
		src, err := format.Source(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("db2go2types: generated %s does not parse: %w\n%s", name, err, buf.Bytes())
		}
		files[name] = src
	}
	return files, nil
}

// Diagram returns a Markdown document with a Mermaid class diagram of the schema.
func Diagram(s *Schema) []byte {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "diagram.tmpl", s); err != nil {
		panic(err) // the template only reads fields of Schema
	}
	return buf.Bytes()
}

type fileData struct {
	Package   string
	NeedsTime bool
	Enums     []enumData
	Tables    []tableData
}

type enumData struct {
	GoName string
	Values []enumValue
}

type enumValue struct {
	Const, Value string
}

type tableData struct {
	GoName    string // Post
	Repo      string // postRepository
	Columns   []columnData
	Params    []columnData // columns without the primary key
	Select    string       // SELECT … FROM "schema"."table"
	Insert    string       // INSERT … RETURNING …
	Update    string       // UPDATE … SET … (no WHERE)
	Returning string       // RETURNING …
	Delete    string
	Count     string
	GetWhere  string // WHERE "id" = $1
}

type columnData struct {
	GoName, GoType string
	Required       bool
}

func newFileData(s *Schema, cfg Config) fileData {
	d := fileData{Package: cfg.Package}
	for _, e := range s.Enums {
		ed := enumData{GoName: goName(e.Name)}
		seen := map[string]int{}
		for _, v := range e.Values {
			name := enumConstName(ed.GoName, v, cfg.EnumNames)
			if seen[name]++; seen[name] > 1 {
				name += strconv.Itoa(seen[name])
			}
			ed.Values = append(ed.Values, enumValue{Const: name, Value: v})
		}
		d.Enums = append(d.Enums, ed)
	}
	for _, t := range s.Tables {
		d.Tables = append(d.Tables, newTableData(s, t, cfg.PrimaryKey))
		for _, c := range t.Columns {
			if strings.Contains(s.goType(c), "time.") {
				d.NeedsTime = true
			}
		}
	}
	return d
}

func newTableData(s *Schema, t Table, pk string) tableData {
	name := goName(t.Name)
	td := tableData{GoName: name, Repo: lowerFirst(name) + "Repository"}
	table := quoteIdent(t.Schema) + "." + quoteIdent(t.Name)

	var selected, params, placeholders, set []string
	for _, c := range t.Columns {
		cd := columnData{GoName: goName(c.Name), GoType: s.goType(c), Required: !c.IsNullable}
		td.Columns = append(td.Columns, cd)
		col := quoteIdent(c.Name)
		if s.isEnumArray(c) {
			selected = append(selected, col+"::text[]")
		} else {
			selected = append(selected, col)
		}
		if c.Name == pk {
			continue
		}
		td.Params = append(td.Params, cd)
		params = append(params, col)
		n := "$" + strconv.Itoa(len(params))
		if s.isEnumArray(c) {
			// pgx cannot encode a slice for an enum array type it does not know; send text[].
			n += "::text[]::" + quoteIdent(t.Schema) + "." + quoteIdent(c.ElementType()) + "[]"
		}
		placeholders = append(placeholders, n)
		set = append(set, col+" = "+n)
	}
	list := strings.Join(selected, ", ")

	td.Returning = " RETURNING " + list
	td.Select = "SELECT " + list + " FROM " + table + " "
	td.GetWhere = "WHERE " + quoteIdent(pk) + " = $1"
	td.Delete = "DELETE FROM " + table + " "
	td.Count = "SELECT COUNT(*) FROM " + table + " "
	if len(params) == 0 {
		td.Insert = "INSERT INTO " + table + " DEFAULT VALUES" + td.Returning
	} else {
		td.Insert = "INSERT INTO " + table + " (" + strings.Join(params, ", ") + ") VALUES (" +
			strings.Join(placeholders, ", ") + ")" + td.Returning
		td.Update = "UPDATE " + table + " SET " + strings.Join(set, ", ") + " "
	}
	return td
}

// quoteIdent quotes a PostgreSQL identifier: user → "user".
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// goString returns a Go string literal, raw when possible.
func goString(s string) string {
	if strings.ContainsAny(s, "`\r") {
		return strconv.Quote(s)
	}
	return "`" + s + "`"
}
