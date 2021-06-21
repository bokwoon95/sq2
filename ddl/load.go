package ddl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/bokwoon95/sq"
)

func (m *Metadata) LoadTable(table sq.Table) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf(fmt.Sprint(r))
			}
		}
	}()
	if table == nil {
		return fmt.Errorf("ddl: table is nil")
	}
	tableValue := reflect.ValueOf(table)
	tableType := tableValue.Type()
	if tableType.Kind() != reflect.Struct {
		return fmt.Errorf("ddl: table is not a struct")
	}
	if tableValue.NumField() == 0 {
		return fmt.Errorf("ddl: table is empty struct")
	}
	genericTable, ok := tableValue.Field(0).Interface().(sq.GenericTable)
	if !ok {
		return fmt.Errorf("ddl: first field of table struct is not an embedded sq.GenericTable")
	}
	if !tableType.Field(0).Anonymous {
		return fmt.Errorf("ddl: first field of table struct is not an embedded sq.GenericTable")
	}
	var schema Schema
	if i := m.CachedSchemaIndex(genericTable.TableSchema); i >= 0 {
		schema = m.Schemas[i]
		defer func(i int) { m.Schemas[i] = schema }(i)
	} else {
		schema = NewSchema(genericTable.TableSchema)
		defer func() { m.AppendSchema(schema) }()
	}
	var tbl Table
	if genericTable.TableName == "" {
		return fmt.Errorf("ddl: table name is empty")
	}
	if i := schema.CachedTableIndex(genericTable.TableName); i >= 0 {
		tbl = schema.Tables[i]
		defer func(i int) { schema.Tables[i] = tbl }(i)
	} else {
		tbl = NewTable(schema.SchemaName, genericTable.TableName)
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
				return fmt.Errorf("ddl: %s: %s", qualifiedTable, err.Error())
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
			err = tbl.LoadConstraint(PRIMARY_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("ddl: %s: %s", qualifiedTable, err.Error())
			}
		case "references":
			err = tbl.LoadConstraint(FOREIGN_KEY, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("ddl: %s: %s", qualifiedTable, err.Error())
			}
		case "unique":
			err = tbl.LoadConstraint(UNIQUE, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("ddl: %s: %s", qualifiedTable, err.Error())
			}
		case "check":
			err = tbl.LoadConstraint(CHECK, tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("ddl: %s: %s", qualifiedTable, err.Error())
			}
		case "index":
			err = tbl.LoadIndex(tbl.TableSchema, tbl.TableName, nil, modifier[1])
			if err != nil {
				return fmt.Errorf("ddl: %s: %s", qualifiedTable, err.Error())
			}
		default:
			return fmt.Errorf("ddl: %s: unknown modifier '%s'", qualifiedTable, modifier[0])
		}
	}
	for i := 1; i < tableValue.NumField(); i++ {
		field, ok := tableValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		columnName := field.GetName()
		if columnName == "" {
			return fmt.Errorf("ddl: table %s field #%d has no name", genericTable.TableName, i)
		}
		columnType := defaultColumnType(m.Dialect, field)
		config := tableType.Field(i).Tag.Get("ddl")
		err := tbl.LoadColumn(m.Dialect, columnName, columnType, config)
		if err != nil {
			return err
		}
	}
	ddlTable, ok := table.(DDLer)
	if !ok {
		return nil
	}
	t := &T{dialect: m.Dialect, tbl: &tbl}
	ddlTable.DDL(m.Dialect, t)
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
	return nil
}

func (tbl *Table) LoadColumn(dialect, columnName, columnType, config string) error {
	qualifiedColumn := tbl.TableSchema + "." + tbl.TableName + "." + columnName
	if tbl.TableSchema == "" {
		qualifiedColumn = qualifiedColumn[1:]
	}
	modifiers, _, err := lexModifiers(config)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
	}
	var col Column
	if i := tbl.CachedColumnIndex(columnName); i >= 0 {
		col = tbl.Columns[i]
		defer func(i int) { tbl.Columns[i] = col }(i)
	} else {
		col = Column{
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			ColumnName:  columnName,
			ColumnType:  columnType,
		}
		defer func() { tbl.AppendColumn(col) }()
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "type":
			col.ColumnType = modifier[1]
		case "autoincrement":
			col.Autoincrement = true
		case "identity":
			col.Identity = BY_DEFAULT_AS_IDENTITY
		case "alwaysidentity":
			col.Identity = ALWAYS_AS_IDENTITY
		case "notnull":
			col.IsNotNull = true
		case "onupdatecurrenttimestamp":
			col.OnUpdateCurrentTimestamp = true
		case "generated":
			col.GeneratedExpr = modifier[1]
		case "stored":
			col.GeneratedExprStored = true
		case "virtual":
			col.GeneratedExprStored = false
		case "collate":
			col.CollationName = modifier[1]
		case "default":
			if len(modifier[1]) >= 2 && modifier[1][0] == '\'' && modifier[1][len(modifier[1])-1] == '\'' {
				col.ColumnDefault = modifier[1]
			} else if strings.EqualFold(modifier[1], "TRUE") ||
				strings.EqualFold(modifier[1], "FALSE") ||
				strings.EqualFold(modifier[1], "CURRENT_DATE") ||
				strings.EqualFold(modifier[1], "CURRENT_TIME") ||
				strings.EqualFold(modifier[1], "CURRENT_TIMESTAMP") {
				col.ColumnDefault = modifier[1]
			} else if _, err := strconv.ParseInt(modifier[1], 10, 64); err == nil {
				col.ColumnDefault = modifier[1]
			} else if _, err := strconv.ParseFloat(modifier[1], 64); err == nil {
				col.ColumnDefault = modifier[1]
			} else if dialect == sq.DialectPostgres {
				col.ColumnDefault = modifier[1]
			} else {
				col.ColumnDefault = "(" + modifier[1] + ")"
			}
		case "ignore":
			col.Ignore = true
		case "primarykey":
			err = tbl.LoadConstraint(PRIMARY_KEY, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "references":
			err = tbl.LoadConstraint(FOREIGN_KEY, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "unique":
			err = tbl.LoadConstraint(UNIQUE, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "check":
			err = tbl.LoadConstraint(CHECK, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "index":
			err = tbl.LoadIndex(col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedColumn, modifier[0])
		}
	}
	return nil
}

func (tbl *Table) LoadConstraint(constraintType, tableSchema, tableName string, columns []string, config string) error {
	value, modifiers, modifierIndex, err := lexValue(config)
	if err != nil {
		return err
	}
	var constraintName string
	if constraintType == PRIMARY_KEY || constraintType == UNIQUE || constraintType == CHECK {
		constraintName = value
	}
	if constraintType == FOREIGN_KEY {
		if i, ok := modifierIndex["name"]; ok {
			constraintName = modifiers[i][1]
		}
	}
	if i, ok := modifierIndex["cols"]; ok {
		columns = strings.Split(modifiers[i][1], ",")
	}
	if constraintName == "." {
		constraintName = ""
	}
	if constraintName == "" && len(columns) > 0 {
		constraintName = generateName(constraintType, tableName, columns...)
	}
	var constraint Constraint
	if i := tbl.CachedConstraintIndex(constraintName); i >= 0 {
		constraint = tbl.Constraints[i]
		constraint.ConstraintType = constraintType
		defer func(i int) { tbl.Constraints[i] = constraint }(i)
	} else {
		constraint = Constraint{
			ConstraintSchema: tableSchema,
			ConstraintName:   constraintName,
			ConstraintType:   constraintType,
			TableSchema:      tableSchema,
			TableName:        tableName,
			Columns:          columns,
		}
		defer func() { tbl.AppendConstraint(constraint) }()
	}
	if constraintType == FOREIGN_KEY {
		switch parts := strings.SplitN(value, ".", 3); len(parts) {
		case 1:
			constraint.ReferencesTable = parts[0]
		case 2:
			constraint.ReferencesTable = parts[0]
			constraint.ReferencesColumns = strings.Split(parts[1], ",")
		case 3:
			constraint.ReferencesSchema = parts[0]
			constraint.ReferencesTable = parts[1]
			constraint.ReferencesColumns = strings.Split(parts[2], ",")
		}
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "schema":
			constraint.ConstraintSchema = modifier[1]
		case "name", "cols":
			continue
		case "onupdate":
			switch modifier[1] {
			case "cascade":
				constraint.OnUpdate = CASCADE
			case "restrict":
				constraint.OnUpdate = RESTRICT
			case "noaction":
				constraint.OnUpdate = NO_ACTION
			case "setnull":
				constraint.OnUpdate = SET_NULL
			case "setdefault":
				constraint.OnUpdate = SET_DEFAULT
			default:
				return fmt.Errorf("unknown value '%s' for 'references.onupdate' modifier", modifier[1])
			}
		case "ondelete":
			switch modifier[1] {
			case "cascade":
				constraint.OnDelete = CASCADE
			case "restrict":
				constraint.OnDelete = RESTRICT
			case "noaction":
				constraint.OnDelete = NO_ACTION
			case "setnull":
				constraint.OnDelete = SET_NULL
			case "setdefault":
				constraint.OnDelete = SET_DEFAULT
			default:
				return fmt.Errorf("unknown value '%s' for 'references.ondelete' modifier", modifier[1])
			}
		case "check":
			constraint.CheckExpr = modifier[1]
		case "deferrable":
			constraint.IsDeferrable = true
		case "deferred":
			constraint.IsInitiallyDeferred = true
		default:
			return fmt.Errorf("invalid modifier 'check.%s'", modifier[0])
		}
	}
	return nil
}

func (tbl *Table) LoadIndex(tableSchema, tableName string, columns []string, config string) error {
	indexName, modifiers, modifierIndex, err := lexValue(config)
	if err != nil {
		return err
	}
	if i, ok := modifierIndex["cols"]; ok {
		columns = strings.Split(modifiers[i][1], ",")
	}
	if indexName == "." {
		indexName = ""
	}
	if indexName == "" && len(columns) > 0 {
		indexName = generateName(INDEX, tableName, columns...)
	}
	var index Index
	if i := tbl.CachedIndexIndex(indexName); i >= 0 {
		index = tbl.Indices[i]
		defer func(i int) { tbl.Indices[i] = index }(i)
	} else {
		index = Index{
			IndexSchema: tableSchema,
			IndexName:   indexName,
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			Columns:     columns,
		}
		defer func() { tbl.AppendIndex(index) }()
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "schema":
			index.IndexSchema = modifier[1]
		case "cols":
			continue
		case "unique":
			index.IsUnique = true
		case "using":
			index.IndexType = strings.ToUpper(modifier[1])
		case "where":
			index.Predicate = modifier[1]
		case "include":
			index.Include = strings.Split(modifier[1], ",")
		default:
			return fmt.Errorf("invalid modifier 'index.%s'", modifier[0])
		}
	}
	return nil
}
