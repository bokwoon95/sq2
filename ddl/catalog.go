package ddl

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/bokwoon95/sq"
)

type Catalog struct {
	Dialect       string      `json:",omitempty"`
	CatalogName   string      `json:",omitempty"`
	VersionString string      `json:",omitempty"`
	VersionNum    [2]int      `json:",omitempty"`
	DefaultSchema string      `json:",omitempty"`
	Extensions    [][2]string `json:",omitempty"`
	Schemas       []*Schema   `json:",omitempty"`
	schemaCache   map[string]int
}

func (c *Catalog) CachedSchemaPosition(schemaName string) (schemaPosition int) {
	schemaPosition, ok := c.schemaCache[schemaName]
	if !ok {
		return -1
	}
	if schemaPosition < 0 || schemaPosition >= len(c.Schemas) || c.Schemas[schemaPosition].SchemaName != schemaName {
		delete(c.schemaCache, schemaName)
		return -1
	}
	return schemaPosition
}

func (c *Catalog) AppendSchema(schema *Schema) (schemaPosition int) {
	c.Schemas = append(c.Schemas, schema)
	if c.schemaCache == nil {
		c.schemaCache = make(map[string]int)
	}
	schemaPosition = len(c.Schemas) - 1
	c.schemaCache[schema.SchemaName] = schemaPosition
	return schemaPosition
}

func (c *Catalog) RefreshSchemaCache() {
	if c.schemaCache == nil && len(c.Schemas) > 0 {
		c.schemaCache = make(map[string]int)
	}
	for i, schema := range c.Schemas {
		c.schemaCache[schema.SchemaName] = i
	}
}

func (c *Catalog) loadTable(table sq.SchemaTable) error {
	if table == nil {
		return fmt.Errorf("table is nil")
	}
	tableSchema, tableName := table.GetSchema(), table.GetName()
	if tableName == "" {
		return fmt.Errorf("table name is empty")
	}
	var schema *Schema
	if n := c.CachedSchemaPosition(tableSchema); n >= 0 {
		schema = c.Schemas[n]
	} else {
		schema = &Schema{SchemaName: tableSchema}
		c.AppendSchema(schema)
	}
	var tbl *Table
	if n := schema.CachedTablePosition(tableName); n >= 0 {
		tbl = schema.Tables[n]
	} else {
		tbl = &Table{
			TableSchema: tableSchema,
			TableName:   tableName,
		}
		schema.AppendTable(tbl)
	}
	return tbl.LoadTable(c.Dialect, table)
}

func (c *Catalog) loadDDLView(ddlView DDLView) error {
	if ddlView == nil {
		return fmt.Errorf("view is nil")
	}
	viewSchema, viewName := ddlView.GetSchema(), ddlView.GetName()
	if viewName == "" {
		return fmt.Errorf("table name is empty")
	}
	var schema *Schema
	if n := c.CachedSchemaPosition(viewSchema); n >= 0 {
		schema = c.Schemas[n]
	} else {
		schema = &Schema{SchemaName: viewSchema}
		c.AppendSchema(schema)
	}
	var view *View
	if n := schema.CachedViewPosition(viewName); n >= 0 {
		view = schema.Views[n]
	} else {
		view = &View{
			ViewSchema: viewSchema,
			ViewName:   viewName,
		}
		schema.AppendView(view)
	}
	return view.LoadDDLView(c.Dialect, ddlView)
}

func (c *Catalog) loadFunction(function Function) error {
	if function.FunctionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	var schema *Schema
	if n := c.CachedSchemaPosition(function.FunctionSchema); n >= 0 {
		schema = c.Schemas[n]
	} else {
		schema = &Schema{SchemaName: function.FunctionSchema}
		c.AppendSchema(schema)
	}
	schema.Functions = append(schema.Functions, &function)
	return nil
}

type CatalogOption func(*Catalog) error

func NewCatalog(dialect string, opts ...CatalogOption) (Catalog, error) {
	catalog := Catalog{Dialect: dialect}
	for _, opt := range opts {
		err := opt(&catalog)
		if err != nil {
			return catalog, err
		}
	}
	return catalog, nil
}

func WithDB(db sq.DB) CatalogOption {
	return func(c *Catalog) error {
		return nil
	}
}

func WithExtensions(extensions ...string) CatalogOption {
	return func(c *Catalog) error {
		for _, extension := range extensions {
			ext := [2]string{extension, ""}
			if i := strings.IndexByte(extension, '@'); i >= 0 {
				ext[0], ext[1] = extension[:i], strings.TrimSpace(extension[i+1:])
			}
			ext[0] = strings.TrimSpace(ext[0])
			c.Extensions = append(c.Extensions, ext)
		}
		return nil
	}
}

func WithTables(tables ...sq.SchemaTable) CatalogOption {
	return func(c *Catalog) error {
		for i, table := range tables {
			err := c.loadTable(table)
			if err != nil {
				return fmt.Errorf("WithTables table #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithDDLViews(ddlViews ...DDLView) CatalogOption {
	return func(c *Catalog) error {
		for i, ddlView := range ddlViews {
			err := c.loadDDLView(ddlView)
			if err != nil {
				return fmt.Errorf("WithDDLViews view #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithFunctions(functions ...Function) CatalogOption {
	return func(c *Catalog) error {
		for i, function := range functions {
			err := function.populateFunctionInfo(c.Dialect)
			if err != nil {
				return fmt.Errorf("WithFunctions function #%d: %w", i+1, err)
			}
			err = c.loadFunction(function)
			if err != nil {
				return fmt.Errorf("WithFunctions function #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithFunctionFiles(fsys fs.FS, filenames ...string) CatalogOption {
	return func(c *Catalog) error {
		for _, filename := range filenames {
			b, err := fs.ReadFile(fsys, filename)
			if err != nil {
				return fmt.Errorf("WithFunctionFiles file %s: %w", filename, err)
			}
			function := Function{SQL: string(b)}
			err = function.populateFunctionInfo(c.Dialect)
			if err != nil {
				return fmt.Errorf("WithFunctionFiles file %s: %w", filename, err)
			}
			err = c.loadFunction(function)
			if err != nil {
				return fmt.Errorf("WithFunctionFiles file %s: %w", filename, err)
			}
		}
		return nil
	}
}

func (c *Catalog) WriteStructs(w io.Writer) error {
	return nil
}
