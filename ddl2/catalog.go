package ddl2

import (
	"fmt"
	"io/fs"
	"reflect"

	"github.com/bokwoon95/sq"
)

type Catalog struct {
	Dialect         string   `json:",omitempty"`
	CatalogName     string   `json:",omitempty"`
	VersionString   string   `json:",omitempty"`
	VersionNums     []int    `json:",omitempty"`
	GeneratedFromDB bool     `json:",omitempty"`
	DefaultSchema   string   `json:",omitempty"`
	Schemas         []Schema `json:",omitempty"`
	schemasCache    map[string]int
}

func (c *Catalog) CachedSchemaPosition(schemaName string) (schemaPosition int) {
	schemaPosition, ok := c.schemasCache[schemaName]
	if !ok || schemaPosition < 0 || schemaPosition >= len(c.Schemas) {
		delete(c.schemasCache, schemaName)
		return -1
	}
	if c.Schemas[schemaPosition].SchemaName != schemaName {
		delete(c.schemasCache, schemaName)
		return -1
	}
	return schemaPosition
}

func (c *Catalog) AppendSchema(schema Schema) (schemaPosition int) {
	c.Schemas = append(c.Schemas, schema)
	if c.schemasCache == nil {
		c.schemasCache = make(map[string]int)
	}
	schemaPosition = len(c.Schemas) - 1
	c.schemasCache[schema.SchemaName] = schemaPosition
	return schemaPosition
}

func (c *Catalog) RefreshSchemasCache() {
	for i, schema := range c.Schemas {
		if c.schemasCache == nil {
			c.schemasCache = make(map[string]int)
		}
		c.schemasCache[schema.SchemaName] = i
	}
}

type CatalogOption func(*Catalog) error

// TODO: implement WithDB
func WithDB(db sq.DB) CatalogOption {
	return func(c *Catalog) error {
		return nil
	}
}

func WithTables(tables ...sq.SchemaTable) CatalogOption {
	return func(c *Catalog) error {
		var err error
		for i, table := range tables {
			err = c.loadTable(table)
			if err != nil {
				return fmt.Errorf("WithTables: table #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithDDLViews(ddlViews ...DDLView) CatalogOption {
	return func(c *Catalog) error {
		var err error
		for i, ddlView := range ddlViews {
			err = c.loadDDLView(ddlView)
			if err != nil {
				return fmt.Errorf("view #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithFunctions(functions ...Function) CatalogOption {
	return func(c *Catalog) error {
		var err error
		for i, function := range functions {
			if function.FunctionName == "" {
				function.FunctionSchema, function.FunctionName, err = getFunctionInfo(c.Dialect, function.SQL)
				if err != nil {
					return fmt.Errorf("function #%d: %w", i+1, err)
				}
			}
			err = c.loadFunction(function)
			if err != nil {
				return fmt.Errorf("function #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

// Maybe I need a different function category, called 'independent functions'.
// nearform/temporal_table's versioning() plpgsql function and the
// last_update_trg() function are examples of independent functions.
// 'Function' and 'DependentFunction'

// TODO: the problem with this is that it provides no way for the user to indicate that the function contains a table.
func WithFunctionFiles(fsys fs.FS, filenames ...string) CatalogOption {
	return func(c *Catalog) error {
		var err error
		var b []byte
		for i, filename := range filenames {
			b, err = fs.ReadFile(fsys, filename)
			if err != nil {
				return fmt.Errorf("file #%d: %w", i+1, err)
			}
			function := Function{SQL: string(b)}
			function.FunctionSchema, function.FunctionName, err = getFunctionInfo(c.Dialect, function.SQL)
			if err != nil {
				return fmt.Errorf("file #%d: %w", i+1, err)
			}
			err = c.loadFunction(function)
			if err != nil {
				return fmt.Errorf("file #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func NewCatalog(dialect string, opts ...CatalogOption) (Catalog, error) {
	catalog := Catalog{Dialect: dialect}
	var err error
	for _, opt := range opts {
		err = opt(&catalog)
		if err != nil {
			return catalog, err
		}
	}
	return catalog, nil
}

func (c *Catalog) loadTable(table sq.SchemaTable) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("panic: " + fmt.Sprint(r))
			}
		}
	}()
	if table == nil {
		return fmt.Errorf("table is nil")
	}
	tableValue := reflect.ValueOf(table)
	tableType := tableValue.Type()
	if tableType.Kind() != reflect.Struct {
		return fmt.Errorf("table is not a struct")
	}
	if tableValue.NumField() == 0 {
		return fmt.Errorf("table is empty struct")
	}
	tableinfo, ok := tableValue.Field(0).Interface().(sq.TableInfo)
	if !ok {
		return fmt.Errorf("first field of table struct is not an embedded sq.TableInfo")
	}
	if !tableType.Field(0).Anonymous {
		return fmt.Errorf("first field of table struct is not an embedded sq.TableInfo")
	}
	var schema Schema
	if n := c.CachedSchemaPosition(tableinfo.TableSchema); n >= 0 {
		schema = c.Schemas[n]
		defer func() { c.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: tableinfo.TableSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	var tbl Table
	if tableinfo.TableName == "" {
		return fmt.Errorf("table name is empty")
	}
	if n := schema.CachedTablePosition(tableinfo.TableName); n >= 0 {
		tbl = schema.Tables[n]
		defer func() { schema.Tables[n] = tbl }()
	} else {
		tbl = Table{TableSchema: tableinfo.TableSchema, TableName: tableinfo.TableName}
		defer func() { schema.AppendTable(tbl) }()
	}
	qualifiedTable := tbl.TableSchema + "." + tbl.TableName
	if tbl.TableSchema == "" {
		qualifiedTable = qualifiedTable[1:]
	}
	tableModifiers := tableType.Field(0).Tag.Get("ddl")
	modifiers, _, err := lexModifiers(tableModifiers)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "virtual":
			virtualTable, modifiers, _, err := lexValue(modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
			tbl.VirtualTable = virtualTable
			for _, modifier := range modifiers {
				virtualTableArg := modifier[0]
				if modifier[1] != "" {
					virtualTableArg += "=" + modifier[1]
				}
				tbl.VirtualTableArgs = append(tbl.VirtualTableArgs, virtualTableArg)
			}
		case "primarykey":
			err = tbl.LoadConstraintConfig(PRIMARY_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "references":
			err = tbl.LoadConstraintConfig(FOREIGN_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "unique":
			err = tbl.LoadConstraintConfig(UNIQUE, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "check":
			err = tbl.LoadConstraintConfig(CHECK, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		case "index":
			err = tbl.LoadIndexConfig(tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedTable, modifier[0])
		}
	}
	for i := 1; i < tableValue.NumField(); i++ {
		field, ok := tableValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		columnName := field.GetName()
		if columnName == "" {
			return fmt.Errorf("table %s field #%d has no name", tableinfo.TableName, i)
		}
		columnType := defaultColumnType(c.Dialect, field)
		config := tableType.Field(i).Tag.Get("ddl")
		err := tbl.LoadColumnConfig(c.Dialect, columnName, columnType, config)
		if err != nil {
			return err
		}
	}
	defer func() {
		for _, constraint := range tbl.Constraints {
			if len(constraint.Columns) != 1 {
				continue
			}
			n := tbl.CachedColumnPosition(constraint.Columns[0])
			if n < 0 {
				continue
			}
			switch constraint.ConstraintType {
			case PRIMARY_KEY:
				tbl.Columns[n].IsPrimaryKey = true
			case UNIQUE:
				tbl.Columns[n].IsUnique = true
			}
		}
	}()
	ddlTable, ok := table.(DDLTable)
	if !ok {
		return nil
	}
	t := &T{dialect: c.Dialect, tbl: &tbl}
	ddlTable.DDL(c.Dialect, t)
	return nil
}

func (c *Catalog) loadDDLView(ddlView DDLView) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("panic: " + fmt.Sprint(r))
			}
		}
	}()
	if ddlView == nil {
		return fmt.Errorf("view is nil")
	}
	ddlViewValue := reflect.ValueOf(ddlView)
	ddlViewType := ddlViewValue.Type()
	if ddlViewType.Kind() != reflect.Struct {
		return fmt.Errorf("view is not a struct")
	}
	if ddlViewValue.NumField() == 0 {
		return fmt.Errorf("view is empty struct")
	}
	tableinfo, ok := ddlViewValue.Field(0).Interface().(sq.TableInfo)
	if !ok {
		return fmt.Errorf("first field of view struct is not an embedded sq.TableInfo")
	}
	if !ddlViewType.Field(0).Anonymous {
		return fmt.Errorf("first field of view struct is not an embedded sq.TableInfo")
	}
	if tableinfo.TableName == "" {
		return fmt.Errorf("view name is empty")
	}
	v := &V{
		dialect: c.Dialect,
		view: &View{
			ViewSchema: tableinfo.TableSchema,
			ViewName:   tableinfo.TableName,
		},
	}
	for i := 1; i < ddlViewValue.NumField(); i++ {
		field, ok := ddlViewValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		fieldName := field.GetName()
		if fieldName == "" {
			return fmt.Errorf("view struct %s field #%d has no name set for it", ddlViewType.Name(), i)
		}
		v.wantColumns = append(v.wantColumns, fieldName)
	}
	ddlView.DDL(c.Dialect, v)
	var schema Schema
	if n := c.CachedSchemaPosition(v.view.ViewSchema); n >= 0 {
		schema = c.Schemas[n]
		defer func() { c.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: v.view.ViewSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	if n := schema.CachedViewPosition(v.view.ViewName); n >= 0 {
		schema.Views[n] = *v.view
	} else {
		schema.AppendView(*v.view)
	}
	return nil
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
