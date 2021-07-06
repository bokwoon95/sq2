package ddl3

import (
	"fmt"
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

func NewCatalog(dialect string) Catalog {
	return Catalog{Dialect: dialect}
}

func (c *Catalog) LoadDB(db sq.Queryer) error {
	return nil
}

func (c *Catalog) LoadTables(tables ...sq.SchemaTable) error {
	var err error
	for i, table := range tables {
		err = c.loadTable(table)
		if err != nil {
			return fmt.Errorf("table #%d: %w", i+1, err)
		}
	}
	return nil
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

func (c *Catalog) LoadDDLViews(ddlViews ...DDLView) error {
	var err error
	for i, ddlView := range ddlViews {
		err = c.loadDDLView(ddlView)
		if err != nil {
			return fmt.Errorf("view #%d: %w", i+1, err)
		}
	}
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

func (c *Catalog) LoadFunctions(functions ...Function) error {
	var err error
	for i, function := range functions {
		err = c.loadFunction(function)
		if err != nil {
			return fmt.Errorf("function #%d: %w", i+1, err)
		}
	}
	return nil
}

func (c *Catalog) loadFunction(function Function) error {
	return nil
}
