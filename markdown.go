package autodoc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
)

// NewMarkdown returns a new markdown instance
func NewMarkdown(fname string) (*Markdown, error) {
	w, err := os.OpenFile(fname, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Markdown{Writer: w}, nil
}

// Markdown implements markdown documentation.
type Markdown struct {
	Writer io.WriteCloser
	types  []reflect.Type
}

// Transport returns a http.RoundTripper that wraps the provided
// roundtripper.
//
// If no roundtripper is provided, the default transport is used.
func (m *Markdown) Transport(tr http.RoundTripper) *TransportMarkdownRecorder {
	return &TransportMarkdownRecorder{Writer: m.Writer, Underlying: tr}
}

// RegisterType registers a concrete object type.
//
// This allows modeling a union type using an interface.
func (m *Markdown) RegisterTypes(vs ...interface{}) {
	for _, v := range vs {
		m.types = append(m.types, reflect.TypeOf(v))
	}
}

// WriteStructTable writes the description for a struct as a table.
//
// The table has three columns: `Field`, `type` and `description`.
//
// The name is the exported field name or the name as described in the
// json tag.
//
// If the json `omit_empty` struct tag is set, then the field is
// described as option.  Readonly fields can be tagged using
// `json:"readonly"` flag.  This can also be specified via
// `doc:"readonly"` (`doc` is treated the same as json).
//
// Nested structs are treated as with url-encoding: the names are
// specified via field.subfield (or field[].subfield).
func (m *Markdown) WriteStructTable(v interface{}) error {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	header := `
| Field | Type | Description |
| ----- | ---- | ----------- |
`
	if _, err := m.Writer.Write([]byte(header)); err != nil {
		return err
	}

	return m.writeStructFields("", t)
}

func (m *Markdown) writeStructFields(namePrefix string, v reflect.Type) error {
	for v.Kind() != reflect.Struct {
		v = v.Elem()
	}

	for kk := 0; kk < v.NumField(); kk++ {
		if err := m.writeStructField(namePrefix, v.Field(kk)); err != nil {
			return err
		}
	}

	return nil
}

func (m *Markdown) writeStructField(namePrefix string, f reflect.StructField) error {
	tag, ok := f.Tag.Lookup("doc")
	if !ok {
		tag = f.Tag.Get("json")
	}
	parts := strings.Split(tag, ",")

	name := namePrefix + m.structFieldName(f, parts)
	sType, err := m.structFieldType(f.Type, parts)
	if err != nil {
		return err
	}
	description := m.structDescription(f)
	attribs := m.structFieldAttributes(contains(parts, "readonly"), contains(parts, "omitempty"))

	_, err = fmt.Fprintf(m.Writer, "| %s | %s %s | %s |\n", name, sType, attribs, description)
	if err == nil && (sType == "Object" || sType == "Array") {
		if sType == "Object" {
			sType = "."
		} else {
			sType = "[]."
		}
		return m.writeStructFields(name+sType, f.Type)
	}

	return err
}

func (m *Markdown) structFieldName(f reflect.StructField, parts []string) string {
	if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
		return parts[0]
	}
	// TODO: convert to snake case
	return f.Name
}

func (m *Markdown) structFieldType(t reflect.Type, parts []string) (string, error) {
	switch t.Kind() {
	case reflect.Bool:
		return "bool", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return "number", nil
	case reflect.Array, reflect.Slice:
		return "Array", nil
		// case reflect.Interface TODO union types
	case reflect.Ptr:
		return m.structFieldType(t.Elem(), parts)
	case reflect.String:
		return "string", nil
	case reflect.Struct:
		if t.Name() == "" || contains(parts, "embed") {
			return "Object", nil
		}
		return t.Name(), nil
	}

	return "", fmt.Errorf("unsupported field type %v", t.Name())
}

func (m *Markdown) structFieldAttributes(readonly, optional bool) string {
	result := []string{}
	if readonly {
		result = append(result, "readonly")
	}
	if optional {
		result = append(result, "optional")
	}
	if len(result) == 0 {
		return ""
	}

	return "(" + strings.Join(result, " ") + ")"
}

func (m *Markdown) structDescription(f reflect.StructField) string {
	return f.Tag.Get("help")
}

func contains(array []string, element string) bool {
	for _, elt := range array {
		if elt == element {
			return true
		}
	}
	return false
}
