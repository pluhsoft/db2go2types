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
	"pgType": func(c Column) string {
		if c.IsArray() {
			return c.ElementType() + "[]"
		}
		return c.Type
	},
}).ParseFS(templateFS, "templates/*.tmpl"))

// ModelsFile is the file written to Config.OutputDir.
const ModelsFile = "models.go"

// Render generates the Go source of [ModelsFile]: a string type with constants
// per enum and a struct per table. The source is gofmt-formatted.
// Only Package and EnumNames of cfg are used.
func Render(s *Schema, cfg Config) ([]byte, error) {
	if cfg.Package == "" {
		cfg.Package = defaultPackage
	}
	if err := checkPackage(cfg.Package); err != nil {
		return nil, err
	}
	data, err := newFileData(s, cfg)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "models.tmpl", data); err != nil {
		return nil, fmt.Errorf("db2go2types: render %s: %w", ModelsFile, err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("db2go2types: generated %s does not parse: %w\n%s", ModelsFile, err, buf.Bytes())
	}
	return src, nil
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
	GoName  string
	Columns []columnData
}

type columnData struct {
	GoName, GoType string
	Required       bool
}

func newFileData(s *Schema, cfg Config) (fileData, error) {
	d := fileData{Package: cfg.Package}
	// Go names of enums, their constants and tables share the package scope.
	declared := map[string]string{}
	declare := func(goName, what string) error {
		if prev, ok := declared[goName]; ok {
			return fmt.Errorf("db2go2types: %s and %s both become the Go name %s", prev, what, goName)
		}
		declared[goName] = what
		return nil
	}

	for _, e := range s.Enums {
		ed := enumData{GoName: goName(e.Name)}
		if err := declare(ed.GoName, "enum "+e.Name); err != nil {
			return d, err
		}
		seen := map[string]int{}
		for _, v := range e.Values {
			name := enumConstName(ed.GoName, v, cfg.EnumNames)
			if seen[name]++; seen[name] > 1 {
				name += strconv.Itoa(seen[name])
			}
			if err := declare(name, fmt.Sprintf("value %q of enum %s", v, e.Name)); err != nil {
				return d, err
			}
			ed.Values = append(ed.Values, enumValue{Const: name, Value: v})
		}
		d.Enums = append(d.Enums, ed)
	}

	for _, t := range s.Tables {
		td := tableData{GoName: goName(t.Name)}
		if err := declare(td.GoName, "table "+t.Name); err != nil {
			return d, err
		}
		fields := map[string]string{}
		for _, c := range t.Columns {
			cd := columnData{GoName: goName(c.Name), GoType: s.goType(c), Required: !c.IsNullable}
			if prev, ok := fields[cd.GoName]; ok {
				return d, fmt.Errorf("db2go2types: columns %s and %s of table %s both become the field %s",
					prev, c.Name, t.Name, cd.GoName)
			}
			fields[cd.GoName] = c.Name
			if strings.Contains(cd.GoType, "time.") {
				d.NeedsTime = true
			}
			td.Columns = append(td.Columns, cd)
		}
		d.Tables = append(d.Tables, td)
	}
	return d, nil
}
