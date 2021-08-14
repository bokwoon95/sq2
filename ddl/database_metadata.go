package ddl

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/bokwoon95/sq"
)

type DatabaseMetadata struct {
	Dialect        string   `json:",omitempty"`
	VersionNums    []int    `json:",omitempty"`
	DatabaseName   string   `json:",omitempty"`
	CurrentSchema  string   `json:",omitempty"`
	Extensions     []string `json:",omitempty"`
	Schemas        []Schema `json:",omitempty"`
	Comment        string   `json:",omitempty"`
	schemaCache    map[string]int
	extensionCache map[string]int
}

func (dbm *DatabaseMetadata) CachedSchemaPosition(schemaName string) (schemaPosition int) {
	schemaPosition, ok := dbm.schemaCache[schemaName]
	if !ok {
		return -1
	}
	if schemaPosition < 0 || schemaPosition >= len(dbm.Schemas) {
		delete(dbm.schemaCache, schemaName)
		return -1
	}
	schema := dbm.Schemas[schemaPosition]
	if schema.SchemaName != schemaName || schema.Ignore {
		delete(dbm.schemaCache, schemaName)
		return -1
	}
	return schemaPosition
}

func (dbm *DatabaseMetadata) AppendSchema(schema Schema) (schemaPosition int) {
	dbm.Schemas = append(dbm.Schemas, schema)
	if dbm.schemaCache == nil {
		dbm.schemaCache = make(map[string]int)
	}
	schemaPosition = len(dbm.Schemas) - 1
	dbm.schemaCache[schema.SchemaName] = schemaPosition
	return schemaPosition
}

func (dbm *DatabaseMetadata) RefreshSchemaCache() {
	if dbm.schemaCache == nil && len(dbm.Schemas) > 0 {
		dbm.schemaCache = make(map[string]int)
	}
	for n, schema := range dbm.Schemas {
		if schema.Ignore {
			continue
		}
		dbm.schemaCache[schema.SchemaName] = n
	}
}

func (dbm *DatabaseMetadata) CachedExtensionPosition(extension string) (extensionPosition int) {
	if i := strings.IndexByte(extension, '@'); i >= 0 {
		extension = extension[:i]
	}
	extensionPosition, ok := dbm.extensionCache[extension]
	if !ok {
		return -1
	}
	if extensionPosition < 0 || extensionPosition >= len(dbm.Schemas) || !strings.HasPrefix(dbm.Extensions[extensionPosition], extension) {
		delete(dbm.schemaCache, extension)
		return -1
	}
	return extensionPosition
}

func (dbm *DatabaseMetadata) AppendExtension(extension string) (extensionPosition int) {
	dbm.Extensions = append(dbm.Extensions, extension)
	if dbm.extensionCache == nil {
		dbm.extensionCache = make(map[string]int)
	}
	extensionPosition = len(dbm.Extensions) - 1
	if i := strings.IndexByte(extension, '@'); i >= 0 {
		extension = extension[:i]
	}
	dbm.extensionCache[extension] = extensionPosition
	return extensionPosition
}

func (dbm *DatabaseMetadata) RefreshExtensionCache() {
	if dbm.extensionCache == nil && len(dbm.Extensions) > 0 {
		dbm.extensionCache = make(map[string]int)
	}
	for n, extension := range dbm.Extensions {
		if i := strings.IndexByte(extension, '@'); i >= 0 {
			extension = extension[:i]
		}
		dbm.extensionCache[extension] = n
	}
}

func (dbm *DatabaseMetadata) loadTable(table sq.SchemaTable) error {
	if table == nil {
		return fmt.Errorf("table is nil")
	}
	tableSchema, tableName := table.GetSchema(), table.GetName()
	if tableName == "" {
		return fmt.Errorf("table name is empty")
	}
	var schema Schema
	if n := dbm.CachedSchemaPosition(tableSchema); n >= 0 {
		schema = dbm.Schemas[n]
		defer func() { dbm.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: tableSchema}
		defer func() { dbm.AppendSchema(schema) }()
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
	return tbl.LoadTable(dbm.Dialect, table)
}

func (dbm *DatabaseMetadata) loadDDLView(ddlView DDLView) error {
	if ddlView == nil {
		return fmt.Errorf("view is nil")
	}
	viewSchema, viewName := ddlView.GetSchema(), ddlView.GetName()
	if viewName == "" {
		return fmt.Errorf("table name is empty")
	}
	var schema Schema
	if n := dbm.CachedSchemaPosition(viewSchema); n >= 0 {
		schema = dbm.Schemas[n]
		defer func() { dbm.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: viewSchema}
		defer func() { dbm.AppendSchema(schema) }()
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
	return view.LoadDDLView(dbm.Dialect, ddlView)
}

func (dbm *DatabaseMetadata) loadFunction(function Function) error {
	if function.FunctionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	var schema Schema
	if n := dbm.CachedSchemaPosition(function.FunctionSchema); n >= 0 {
		schema = dbm.Schemas[n]
		defer func() { dbm.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: function.FunctionSchema}
		defer func() { dbm.AppendSchema(schema) }()
	}
	schema.Functions = append(schema.Functions, function)
	return nil
}

type DatabaseMetadataOption func(*DatabaseMetadata) error

func NewDatabaseMetadata(dialect string, opts ...DatabaseMetadataOption) (DatabaseMetadata, error) {
	dbMetadata := DatabaseMetadata{Dialect: dialect}
	for _, opt := range opts {
		err := opt(&dbMetadata)
		if err != nil {
			return dbMetadata, err
		}
	}
	return dbMetadata, nil
}

func WithDB(db sq.DB, defaultFilter *Filter) DatabaseMetadataOption {
	return func(c *DatabaseMetadata) error {
		dbi, err := NewDatabaseIntrospector(c.Dialect, db, defaultFilter)
		if err != nil {
			return fmt.Errorf("NewDatabaseIntrospector: %w", err)
		}
		ctx := context.Background()
		c.VersionNums, err = dbi.GetVersionNums(ctx)
		if err != nil {
			return fmt.Errorf("GetVersionNums: %w", err)
		}
		c.DatabaseName, err = dbi.GetDatabaseName(ctx)
		if err != nil {
			return fmt.Errorf("GetDatabaseName: %w", err)
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
		constraintNames := make(map[[3]string]struct{})
		for _, constraint := range constraints {
			n1 := c.CachedSchemaPosition(constraint.TableSchema)
			if n1 < 0 {
				continue
			}
			n2 := c.Schemas[n1].CachedTablePosition(constraint.TableName)
			if n2 < 0 {
				continue
			}
			constraintNames[[3]string{constraint.TableSchema, constraint.TableName, constraint.ConstraintName}] = struct{}{}
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
				if _, ok := constraintNames[[3]string{index.TableSchema, index.TableName, index.IndexName}]; ok {
					continue
				}
				if n2 := c.Schemas[n1].CachedTablePosition(index.TableName); n2 >= 0 {
					c.Schemas[n1].Tables[n2].AppendIndex(index)
				} else if n3 := c.Schemas[n1].CachedViewPosition(index.TableName); n3 >= 0 {
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

func WithExtensions(extensions ...string) DatabaseMetadataOption {
	return func(c *DatabaseMetadata) error {
		c.Extensions = extensions
		return nil
	}
}

func WithTables(tables ...sq.SchemaTable) DatabaseMetadataOption {
	return func(c *DatabaseMetadata) error {
		for i, table := range tables {
			err := c.loadTable(table)
			if err != nil {
				return fmt.Errorf("WithTables table #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithDDLViews(ddlViews ...DDLView) DatabaseMetadataOption {
	return func(c *DatabaseMetadata) error {
		for i, ddlView := range ddlViews {
			err := c.loadDDLView(ddlView)
			if err != nil {
				return fmt.Errorf("WithDDLViews view #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithFunctions(functions ...Function) DatabaseMetadataOption {
	return func(c *DatabaseMetadata) error {
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

func (dbm *DatabaseMetadata) WriteStructs(w io.Writer) error {
	return nil
}
