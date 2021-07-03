package ddl2

import (
	"fmt"
	"reflect"

	"github.com/bokwoon95/sq"
)

type Metadata struct {
	Dialect         string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
	Schemas         []Schema
	schemasCache    map[string]int
}

func NewMetadata(dialect string) Metadata {
	return Metadata{Dialect: dialect}
}

func (m *Metadata) LoadDB() error {
	return nil
}

func (m *Metadata) LoadTables(tables ...sq.SchemaTable) error {
	var err error
	for i, table := range tables {
		err = m.LoadTable(table)
		if err != nil {
			qualifiedTableName := table.GetSchema() + "." + table.GetName()
			if qualifiedTableName[0] == '.' {
				qualifiedTableName = qualifiedTableName[1:]
			}
			return fmt.Errorf("table #%d %s: %w", i+1, qualifiedTableName, err)
		}
	}
	return nil
}

func (m *Metadata) LoadTable(table sq.SchemaTable) (err error) {
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
	if i := m.CachedSchemaIndex(tableinfo.TableSchema); i >= 0 {
		schema = m.Schemas[i]
		defer func(i int) { m.Schemas[i] = schema }(i)
	} else {
		schema = Schema{SchemaName: tableinfo.TableSchema}
		defer func() { m.AppendSchema(schema) }()
	}
	var tbl Table
	if tableinfo.TableName == "" {
		return fmt.Errorf("table name is empty")
	}
	if i := schema.CachedTableIndex(tableinfo.TableName); i >= 0 {
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
		columnType := defaultColumnType(m.Dialect, field)
		config := tableType.Field(i).Tag.Get("ddl")
		err := tbl.LoadColumnConfig(m.Dialect, columnName, columnType, config)
		if err != nil {
			return err
		}
	}
	ddlTable, ok := table.(DDLer)
	defer func() {
		for _, constraint := range tbl.Constraints {
			if len(constraint.Columns) != 1 {
				continue
			}
			columnIndex := tbl.CachedColumnIndex(constraint.Columns[0])
			if columnIndex < 0 {
				continue
			}
			switch constraint.ConstraintType {
			case PRIMARY_KEY:
				tbl.Columns[columnIndex].IsPrimaryKey = true
			case UNIQUE:
				tbl.Columns[columnIndex].IsUnique = true
			}
		}
	}()
	if !ok {
		return nil
	}
	t := &T{dialect: m.Dialect, tbl: &tbl}
	ddlTable.DDL(m.Dialect, t)
	return nil
}

func (m *Metadata) CachedSchemaIndex(schemaName string) (schemaIndex int) {
	if m == nil {
		return -1
	}
	schemaIndex, ok := m.schemasCache[schemaName]
	if !ok || schemaIndex < 0 || schemaIndex >= len(m.Schemas) {
		delete(m.schemasCache, schemaName)
		return -1
	}
	if m.Schemas[schemaIndex].SchemaName != schemaName {
		delete(m.schemasCache, schemaName)
		return -1
	}
	return schemaIndex
}

func (m *Metadata) AppendSchema(schema Schema) (schemaIndex int) {
	if m == nil {
		return -1
	}
	m.Schemas = append(m.Schemas, schema)
	if m.schemasCache == nil {
		m.schemasCache = make(map[string]int)
	}
	schemaIndex = len(m.Schemas) - 1
	m.schemasCache[schema.SchemaName] = schemaIndex
	return schemaIndex
}

func (m *Metadata) RefreshSchemaCache() {
	if m == nil {
		return
	}
	for i, schema := range m.Schemas {
		if m.schemasCache == nil {
			m.schemasCache = make(map[string]int)
		}
		m.schemasCache[schema.SchemaName] = i
	}
}
