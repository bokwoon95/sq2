package ddl

import (
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"github.com/bokwoon95/sq"
)

type Catalog struct {
	Dialect         string
	CatalogName     string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
	DefaultSchema   string
	Schemas         []Schema
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

func (c *Catalog) loadDDLView(ddlView DDLView) error {
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
	fieldNames := make(map[string]int)
	for i := 1; i < ddlViewValue.NumField(); i++ {
		field, ok := ddlViewValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		fieldName := field.GetName()
		if fieldName == "" {
			return fmt.Errorf("view struct %s field #%d has no name set for it", ddlViewType.Name(), i)
		}
		fieldNames[fieldName]--
	}
	v := &V{}
	query := ddlView.DDL(c.Dialect, v)
	view := View{
		ViewSchema: tableinfo.TableSchema,
		ViewName:   tableinfo.TableName,
	}
	err := view.loadQuery(query, v)
	if err != nil {
		return fmt.Errorf("view %s loading query: %w", view.ViewName, err)
	}
	for _, fieldName := range view.FieldNames {
		fieldNames[fieldName]++
	}
	var missingFields, extraFields []string
	for fieldName, n := range fieldNames {
		if n > 0 {
			extraFields = append(extraFields, fieldName)
		} else if n < 0 {
			missingFields = append(missingFields, fieldName)
		}
	}
	if len(missingFields) > 0 || len(extraFields) > 0 {
		errMsg := fmt.Sprintf("view %s query fields does not match struct fields:", view.ViewName)
		if len(missingFields) > 0 {
			errMsg += fmt.Sprintf(" (missingFields=%s)", strings.Join(missingFields, ", "))
		}
		if len(extraFields) > 0 {
			errMsg += fmt.Sprintf(" (extraFields=%s)", strings.Join(extraFields, ", "))
		}
		return fmt.Errorf(errMsg)
	}
	var schema Schema
	if n := c.CachedSchemaPosition(view.ViewSchema); n >= 0 {
		schema = c.Schemas[n]
		defer func() { c.Schemas[n] = schema }()
	} else {
		schema = Schema{SchemaName: view.ViewSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	if n := schema.CachedViewPosition(view.ViewName); n >= 0 {
		schema.Views[n] = view
	} else {
		schema.AppendView(view)
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
