package ddl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bokwoon95/sq"
)

type Catalog struct {
	Dialect        string   `json:",omitempty"`
	VersionNums    []int    `json:",omitempty"`
	CatalogName    string   `json:",omitempty"`
	CurrentSchema  string   `json:",omitempty"`
	Extensions     []string `json:",omitempty"`
	Schemas        []Schema `json:",omitempty"`
	schemaCache    map[string]int
	extensionCache map[string]int
}

func (c *Catalog) CachedSchemaPosition(schemaName string) (schemaPosition int) {
	schemaPosition, ok := c.schemaCache[schemaName]
	if !ok {
		return -1
	}
	if schemaPosition < 0 || schemaPosition >= len(c.Schemas) {
		delete(c.schemaCache, schemaName)
		return -1
	}
	schema := c.Schemas[schemaPosition]
	if schema.SchemaName != schemaName || schema.Ignore {
		delete(c.schemaCache, schemaName)
		return -1
	}
	return schemaPosition
}

func (c *Catalog) AppendSchema(schema Schema) (schemaPosition int) {
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
	for n, schema := range c.Schemas {
		if schema.Ignore {
			continue
		}
		c.schemaCache[schema.SchemaName] = n
	}
}

func (c *Catalog) CachedExtensionPosition(extension string) (extensionPosition int) {
	if i := strings.IndexByte(extension, '@'); i >= 0 {
		extension = extension[:i]
	}
	extensionPosition, ok := c.extensionCache[extension]
	if !ok {
		return -1
	}
	if extensionPosition < 0 || extensionPosition >= len(c.Schemas) || !strings.HasPrefix(c.Extensions[extensionPosition], extension) {
		delete(c.schemaCache, extension)
		return -1
	}
	return extensionPosition
}

func (c *Catalog) AppendExtension(extension string) (extensionPosition int) {
	c.Extensions = append(c.Extensions, extension)
	if c.extensionCache == nil {
		c.extensionCache = make(map[string]int)
	}
	extensionPosition = len(c.Extensions) - 1
	if i := strings.IndexByte(extension, '@'); i >= 0 {
		extension = extension[:i]
	}
	c.extensionCache[extension] = extensionPosition
	return extensionPosition
}

func (c *Catalog) RefreshExtensionCache() {
	if c.extensionCache == nil && len(c.Extensions) > 0 {
		c.extensionCache = make(map[string]int)
	}
	for n, extension := range c.Extensions {
		if i := strings.IndexByte(extension, '@'); i >= 0 {
			extension = extension[:i]
		}
		c.extensionCache[extension] = n
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
	var schema Schema
	if n := c.CachedSchemaPosition(tableSchema); n >= 0 {
		schema = c.Schemas[n]
		defer func() { c.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: tableSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	var tbl Table
	if n := schema.CachedTablePosition(tableName); n >= 0 {
		tbl = schema.Tables[n]
		defer func() { schema.Tables[n] = tbl }()
	} else {
		tbl = Table{
			TableSchema: tableSchema,
			TableName:   tableName,
		}
		defer func() { schema.AppendTable(tbl) }()
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
	var schema Schema
	if n := c.CachedSchemaPosition(viewSchema); n >= 0 {
		schema = c.Schemas[n]
		defer func() { c.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: viewSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	var view View
	if n := schema.CachedViewPosition(viewName); n >= 0 {
		view = schema.Views[n]
		defer func() { schema.Views[n] = view }()
	} else {
		view = View{
			ViewSchema: viewSchema,
			ViewName:   viewName,
		}
		defer func() { schema.AppendView(view) }()
	}
	return view.LoadDDLView(c.Dialect, ddlView)
}

func (c *Catalog) loadFunction(function Function) error {
	if function.FunctionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	var schema Schema
	if n := c.CachedSchemaPosition(function.FunctionSchema); n >= 0 {
		schema = c.Schemas[n]
		defer func() { c.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: function.FunctionSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	schema.Functions = append(schema.Functions, function)
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

func WithDB(db sq.DB, defaultFilter *Filter) CatalogOption {
	return func(c *Catalog) error {
		dbi, err := NewDatabaseIntrospector(c.Dialect, db, defaultFilter)
		if err != nil {
			return fmt.Errorf("NewDatabaseIntrospector: %w", err)
		}
		ctx := context.Background()
		c.VersionNums, err = dbi.GetVersionNums(ctx)
		if err != nil {
			return fmt.Errorf("GetVersionNums: %w", err)
		}
		c.CatalogName, err = dbi.GetCatalogName(ctx)
		if err != nil {
			return fmt.Errorf("GetCatalogName: %w", err)
		}
		c.CurrentSchema, err = dbi.GetCurrentSchema(ctx)
		if err != nil {
			return fmt.Errorf("GetCurrentSchema: %w", err)
		}
		if c.Dialect == sq.DialectPostgres {
			c.Extensions, err = dbi.GetExtensions(ctx, nil)
			if err != nil {
				return fmt.Errorf("GetExtensions: %w", err)
			}
		}
		tbls, err := dbi.GetTables(ctx, nil)
		if err != nil {
			return fmt.Errorf("GetTables: %w", err)
		}
		schemaTableCount := make(map[string]int)
		for _, tbl := range tbls {
			schemaTableCount[tbl.TableSchema]++
		}
		c.Schemas = make([]Schema, 0, len(schemaTableCount))
		for _, tbl := range tbls {
			n := c.CachedSchemaPosition(tbl.TableSchema)
			if n < 0 {
				n = c.AppendSchema(Schema{
					SchemaName: tbl.TableSchema,
					Tables:     make([]Table, 0, schemaTableCount[tbl.TableSchema]),
				})
			}
			c.Schemas[n].AppendTable(tbl)
		}
		views, err := dbi.GetViews(ctx, nil)
		if err != nil {
			return fmt.Errorf("GetViews: %w", err)
		}
		for _, view := range views {
			n := c.CachedSchemaPosition(view.ViewSchema)
			if n < 0 {
				n = c.AppendSchema(Schema{SchemaName: view.ViewSchema})
			}
			c.Schemas[n].AppendView(view)
		}
		columns, err := dbi.GetColumns(ctx, nil)
		if err != nil {
			return fmt.Errorf("GetColumns: %w", err)
		}
		for _, column := range columns {
			if n1 := c.CachedSchemaPosition(column.TableSchema); n1 >= 0 {
				if n2 := c.Schemas[n1].CachedTablePosition(column.TableName); n2 >= 0 {
					c.Schemas[n1].Tables[n2].AppendColumn(column)
				}
			}
		}
		constraints, err := dbi.GetConstraints(ctx, nil)
		if err != nil {
			return fmt.Errorf("GetConstraints: %w", err)
		}
		for _, constraint := range constraints {
			n1 := c.CachedSchemaPosition(constraint.TableSchema)
			if n1 < 0 {
				continue
			}
			n2 := c.Schemas[n1].CachedTablePosition(constraint.TableName)
			if n2 < 0 {
				continue
			}
			c.Schemas[n1].Tables[n2].AppendConstraint(constraint)
			if len(constraint.Columns) == 1 && (constraint.ConstraintType == PRIMARY_KEY || constraint.ConstraintType == UNIQUE) {
				n3 := c.Schemas[n1].Tables[n2].CachedColumnPosition(constraint.Columns[0])
				if n3 < 0 {
					continue
				}
				switch constraint.ConstraintType {
				case PRIMARY_KEY:
					c.Schemas[n1].Tables[n2].Columns[n3].IsPrimaryKey = true
				case UNIQUE:
					c.Schemas[n1].Tables[n2].Columns[n3].IsUnique = true
				}
			}
		}
		indexes, err := dbi.GetIndexes(ctx, nil)
		if err != nil {
			return fmt.Errorf("GetConstraints: %w", err)
		}
		for _, index := range indexes {
			if n1 := c.CachedSchemaPosition(index.TableSchema); n1 >= 0 {
				if n2 := c.Schemas[n1].CachedTablePosition(index.TableName); n2 >= 0 {
					c.Schemas[n1].Tables[n2].AppendIndex(index)
				}
				if n3 := c.Schemas[n1].CachedViewPosition(index.TableName); n3 >= 0 {
					c.Schemas[n1].Views[n3].AppendIndex(index)
				}
			}
		}
		if dbi.dialect == sq.DialectPostgres {
			functions, err := dbi.GetFunctions(ctx, nil)
			if err != nil {
				return fmt.Errorf("GetFunctions: %w", err)
			}
			for _, function := range functions {
				n1 := c.CachedSchemaPosition(function.FunctionSchema)
				if n1 < 0 {
					n1 = c.AppendSchema(Schema{SchemaName: function.FunctionSchema})
				}
				c.Schemas[n1].AppendFunction(function)
			}
		}
		triggers, err := dbi.GetTriggers(ctx, nil)
		if err != nil {
			return fmt.Errorf("GetTriggers: %w", err)
		}
		for _, trigger := range triggers {
			if n1 := c.CachedSchemaPosition(trigger.TableSchema); n1 >= 0 {
				if n2 := c.Schemas[n1].CachedTablePosition(trigger.TableName); n2 >= 0 {
					c.Schemas[n1].Tables[n2].AppendTrigger(trigger)
				}
				if n3 := c.Schemas[n1].CachedViewPosition(trigger.TableName); n3 >= 0 {
					c.Schemas[n1].Views[n3].AppendTrigger(trigger)
				}
			}
		}
		return nil
	}
}

func WithExtensions(extensions ...string) CatalogOption {
	return func(c *Catalog) error {
		c.Extensions = extensions
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

func (c *Catalog) WriteStructs(w io.Writer) error {
	return nil
}
