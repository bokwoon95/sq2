package ddl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/bokwoon95/sq"
)

type Table struct {
	TableSchema      string       `json:",omitempty"`
	TableName        string       `json:",omitempty"`
	Columns          []Column     `json:",omitempty"`
	Constraints      []Constraint `json:",omitempty"`
	Indexes          []Index      `json:",omitempty"`
	Triggers         []Trigger    `json:",omitempty"`
	VirtualTable     string       `json:",omitempty"`
	VirtualTableArgs []string     `json:",omitempty"`
	columnCache      map[string]int
	constraintCache  map[string]int
	indexCache       map[string]int
	triggerCache     map[string]int
}

func (tbl *Table) CachedColumnPosition(columnName string) (columnPosition int) {
	if columnName == "" {
		return -1
	}
	columnPosition, ok := tbl.columnCache[columnName]
	if !ok {
		return -1
	}
	if columnPosition < 0 || columnPosition >= len(tbl.Columns) || tbl.Columns[columnPosition].ColumnName != columnName {
		delete(tbl.columnCache, columnName)
		return -1
	}
	return columnPosition
}

func (tbl *Table) AppendColumn(column Column) (columnPosition int) {
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnCache == nil {
		tbl.columnCache = make(map[string]int)
	}
	columnPosition = len(tbl.Columns) - 1
	tbl.columnCache[column.ColumnName] = columnPosition
	return columnPosition
}

func (tbl *Table) RefreshColumnCache() {
	if tbl.columnCache == nil && len(tbl.Columns) > 0 {
		tbl.columnCache = make(map[string]int)
	}
	for i, column := range tbl.Columns {
		tbl.columnCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintPosition(constraintName string) (constraintPosition int) {
	if constraintName == "" {
		return -1
	}
	constraintPosition, ok := tbl.constraintCache[constraintName]
	if !ok {
		return -1
	}
	if constraintPosition < 0 || constraintPosition >= len(tbl.Constraints) || tbl.Constraints[constraintPosition].ConstraintName != constraintName {
		delete(tbl.constraintCache, constraintName)
		return -1
	}
	return constraintPosition
}

func (tbl *Table) AppendConstraint(constraint Constraint) (constraintPosition int) {
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintCache == nil {
		tbl.constraintCache = make(map[string]int)
	}
	constraintPosition = len(tbl.Constraints) - 1
	tbl.constraintCache[constraint.ConstraintName] = constraintPosition
	return constraintPosition
}

func (tbl *Table) RefreshConstraintCache() {
	if tbl.constraintCache == nil && len(tbl.Constraints) > 0 {
		tbl.constraintCache = make(map[string]int)
	}
	for i, constraint := range tbl.Constraints {
		tbl.constraintCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexPosition(indexName string) (indexPosition int) {
	if indexName == "" {
		return -1
	}
	indexPosition, ok := tbl.indexCache[indexName]
	if !ok {
		return -1
	}
	if indexPosition < 0 || indexPosition >= len(tbl.Indexes) || tbl.Indexes[indexPosition].IndexName != indexName {
		delete(tbl.indexCache, indexName)
		return -1
	}
	return indexPosition
}

func (tbl *Table) AppendIndex(index Index) (indexPosition int) {
	tbl.Indexes = append(tbl.Indexes, index)
	if tbl.indexCache == nil {
		tbl.indexCache = make(map[string]int)
	}
	indexPosition = len(tbl.Indexes) - 1
	tbl.indexCache[index.IndexName] = indexPosition
	return indexPosition
}

func (tbl *Table) RefreshIndexesCache() {
	if tbl.indexCache == nil && len(tbl.Indexes) > 0 {
		tbl.indexCache = make(map[string]int)
	}
	for i, index := range tbl.Indexes {
		tbl.indexCache[index.IndexName] = i
	}
}

func (tbl *Table) CachedTriggerPosition(triggerName string) (triggerPosition int) {
	if triggerName == "" {
		return -1
	}
	triggerPosition, ok := tbl.triggerCache[triggerName]
	if !ok {
		return -1
	}
	if triggerPosition < 0 || triggerPosition >= len(tbl.Triggers) || tbl.Triggers[triggerPosition].TriggerName != triggerName {
		delete(tbl.triggerCache, triggerName)
		return -1
	}
	return triggerPosition
}

func (tbl *Table) AppendTrigger(trigger Trigger) (triggerPosition int) {
	tbl.Triggers = append(tbl.Triggers, trigger)
	if tbl.triggerCache == nil {
		tbl.triggerCache = make(map[string]int)
	}
	triggerPosition = len(tbl.Triggers) - 1
	tbl.triggerCache[trigger.TriggerName] = triggerPosition
	return triggerPosition
}

func (tbl *Table) RefreshTriggerCache() {
	if tbl.triggerCache == nil && len(tbl.Triggers) > 0 {
		tbl.triggerCache = make(map[string]int)
	}
	for i, trigger := range tbl.Triggers {
		tbl.triggerCache[trigger.TriggerName] = i
	}
}

func (tbl *Table) LoadIndexConfig(tableSchema, tableName string, columns []string, config string) error {
	indexName, modifiers, modifierPositions, err := tokenizeValue(config)
	if err != nil {
		return err
	}
	if n, ok := modifierPositions["cols"]; ok {
		columns = strings.Split(modifiers[n][1], ",")
	}
	if indexName == "." {
		indexName = ""
	}
	if indexName == "" && len(columns) > 0 {
		indexName = generateName(INDEX, tableName, columns...)
	}
	var index Index
	if n := tbl.CachedIndexPosition(indexName); n >= 0 {
		index = tbl.Indexes[n]
		defer func() { tbl.Indexes[n] = index }()
	} else {
		index = Index{
			TableSchema: tbl.TableSchema,
			TableName:   tbl.TableName,
			IndexName:   indexName,
			Columns:     columns,
		}
		defer func() { tbl.AppendIndex(index) }()
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "cols":
			continue
		case "unique":
			index.IsUnique = true
		case "using":
			index.IndexType = strings.ToUpper(modifier[1])
		case "where":
			index.Predicate = modifier[1]
		case "include":
			index.IncludeColumns = strings.Split(modifier[1], ",")
		default:
			return fmt.Errorf("invalid modifier 'index.%s'", modifier[0])
		}
	}
	return nil
}

func (tbl *Table) LoadConstraintConfig(constraintType, tableSchema, tableName string, columns []string, config string) error {
	value, modifiers, modifierPositions, err := tokenizeValue(config)
	if err != nil {
		return err
	}
	var constraintName string
	if constraintType == PRIMARY_KEY || constraintType == UNIQUE || constraintType == CHECK {
		constraintName = value
	}
	if constraintType == FOREIGN_KEY {
		if n, ok := modifierPositions["name"]; ok {
			constraintName = modifiers[n][1]
		}
	}
	if n, ok := modifierPositions["cols"]; ok {
		columns = strings.Split(modifiers[n][1], ",")
	}
	if constraintName == "." {
		constraintName = ""
	}
	if constraintName == "" && len(columns) > 0 {
		constraintName = generateName(constraintType, tableName, columns...)
	}
	var constraint Constraint
	if n := tbl.CachedConstraintPosition(constraintName); n >= 0 {
		constraint = tbl.Constraints[n]
		constraint.ConstraintType = constraintType
		defer func() { tbl.Constraints[n] = constraint }()
	} else {
		constraint = Constraint{
			TableSchema:    tableSchema,
			TableName:      tableName,
			ConstraintName: constraintName,
			ConstraintType: constraintType,
			Columns:        columns,
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
		case "name", "cols":
			continue
		case "onupdate":
			switch modifier[1] {
			case "cascade":
				constraint.UpdateRule = CASCADE
			case "restrict":
				constraint.UpdateRule = RESTRICT
			case "noaction":
				constraint.UpdateRule = NO_ACTION
			case "setnull":
				constraint.UpdateRule = SET_NULL
			case "setdefault":
				constraint.UpdateRule = SET_DEFAULT
			default:
				return fmt.Errorf("unknown value '%s' for 'references.onupdate' modifier", modifier[1])
			}
		case "ondelete":
			switch modifier[1] {
			case "cascade":
				constraint.DeleteRule = CASCADE
			case "restrict":
				constraint.DeleteRule = RESTRICT
			case "noaction":
				constraint.DeleteRule = NO_ACTION
			case "setnull":
				constraint.DeleteRule = SET_NULL
			case "setdefault":
				constraint.DeleteRule = SET_DEFAULT
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

func (tbl *Table) LoadColumnConfig(dialect, columnName, columnType, config string) error {
	qualifiedColumn := tbl.TableSchema + "." + tbl.TableName + "." + columnName
	if tbl.TableSchema == "" {
		qualifiedColumn = qualifiedColumn[1:]
	}
	modifiers, _, err := tokenizeModifiers(config)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
	}
	var col Column
	if n := tbl.CachedColumnPosition(columnName); n >= 0 {
		col = tbl.Columns[n]
		defer func() { tbl.Columns[n] = col }()
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
			err = tbl.LoadConstraintConfig(PRIMARY_KEY, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "references":
			err = tbl.LoadConstraintConfig(FOREIGN_KEY, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "unique":
			err = tbl.LoadConstraintConfig(UNIQUE, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "check":
			err = tbl.LoadConstraintConfig(CHECK, col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		case "index":
			err = tbl.LoadIndexConfig(col.TableSchema, col.TableName, []string{col.ColumnName}, modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedColumn, err.Error())
			}
		default:
			return fmt.Errorf("%s: unknown modifier '%s'", qualifiedColumn, modifier[0])
		}
	}
	return nil
}

func (tbl *Table) LoadTable(dialect string, table sq.SchemaTable) (err error) {
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
	tbl.TableSchema, tbl.TableName = table.GetSchema(), table.GetName()
	if tbl.TableName == "" {
		return fmt.Errorf("table name is empty")
	}
	qualifiedTable := tbl.TableName
	if tbl.TableSchema != "" {
		qualifiedTable = tbl.TableSchema + "." + tbl.TableName
	}
	tableModifiers := tableType.Field(0).Tag.Get("ddl")
	modifiers, _, err := tokenizeModifiers(tableModifiers)
	if err != nil {
		return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
	}
	for _, modifier := range modifiers {
		switch modifier[0] {
		case "virtual":
			virtualTable, submodifiers, _, err := tokenizeValue(modifier[1])
			if err != nil {
				return fmt.Errorf("%s: %s", qualifiedTable, err.Error())
			}
			tbl.VirtualTable = virtualTable
			for _, submodifier := range submodifiers {
				virtualTableArg := submodifier[0]
				if submodifier[1] != "" {
					virtualTableArg += "=" + submodifier[1]
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
	for i := 0; i < tableValue.NumField(); i++ {
		field, ok := tableValue.Field(i).Interface().(sq.Field)
		if !ok {
			continue
		}
		columnName := field.GetName()
		if columnName == "" {
			return fmt.Errorf("table %s field #%d has no name", tbl.TableName, i)
		}
		columnType := defaultColumnType(dialect, field)
		config := tableType.Field(i).Tag.Get("ddl")
		err = tbl.LoadColumnConfig(dialect, columnName, columnType, config)
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
	// ddlTable, ok := table.(DDLTable)
	// if !ok {
	// 	return nil
	// }
	return nil
}
