package ddl3

import (
	"fmt"
	"io/fs"
	"reflect"

	"github.com/bokwoon95/sq"
)

type Catalog struct {
	Dialect         string
	CatalogName     string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
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

// The prime motivation for streamlining the LoadDB and
// LoadTables+LoadDDLViews+LoadFunctions workflow is to encourage people to
// roll their own AutoMigrate, because that's where any sort of real
// configuration power takes place. Don't make loading from db and loading from
// tables/etc so painful that people avoid it because of the error checking
// boilerplate.


type CatalogOption func(*Catalog) error

func WithDB(db sq.Queryer) CatalogOption {
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
				return fmt.Errorf("WithTables table #%d: %w", i+1, err)
			}
		}
		return nil
	}
}

func WithDDLViews(ddlViews ...DDLView) CatalogOption {
	return func(c *Catalog) error {
		return nil
	}
}

// TODO: toy with the idea of taking in a func(fn *Fn) instead
// TODO: figure out how to accomodate WithFunction(Sprintf) and WithFunction(Filef)
// TODO: OH GOD: if I can figure out how to scan the first few characters of a
// CREATE TRIGGER or CREATE FUNCTION command for the schema name and the
// trigger/function name, I can dispense with repeating the trigger or function
// name in application code.
// NOTE: Do I still want Functionf/FunctionFilef? Unlike triggers, functions
// that require extensive templating can be very big, and the args list passed
// to Functionf could be very big and clunky. Not sure if that's what I want
// users to do.
func WithFunctionFile(functionSchema, functionName string, fsys fs.FS, filename string) CatalogOption {
	return func(c *Catalog) error {
		return nil
	}
}

func NewCatalog(dialect string, opts ...CatalogOption) Catalog {
	return Catalog{Dialect: dialect}
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
	if i := c.CachedSchemaPosition(tableinfo.TableSchema); i >= 0 {
		schema = c.Schemas[i]
		defer func(i int) { c.Schemas[i] = schema }(i)
	} else {
		schema = Schema{SchemaName: tableinfo.TableSchema}
		defer func() { c.AppendSchema(schema) }()
	}
	var tbl Table
	if tableinfo.TableName == "" {
		return fmt.Errorf("table name is empty")
	}
	if i := schema.CachedTablePosition(tableinfo.TableName); i >= 0 {
		tbl = schema.Tables[i]
		defer func(i int) { schema.Tables[i] = tbl }(i)
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
			columnPosition := tbl.CachedColumnPosition(constraint.Columns[0])
			if columnPosition < 0 {
				continue
			}
			switch constraint.ConstraintType {
			case PRIMARY_KEY:
				tbl.Columns[columnPosition].IsPrimaryKey = true
			case UNIQUE:
				tbl.Columns[columnPosition].IsUnique = true
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
	v := &V{}
	query := ddlView.DDL(c.Dialect, v)
	gotFields, err := query.GetFetchableFields()
	if err != nil {
		return fmt.Errorf("fetching view fields: %w", err)
	}
	_ = gotFields
	return nil
}

func (c *Catalog) loadFunction(function Function) error {
	return nil
}
