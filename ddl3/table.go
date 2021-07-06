package ddl3

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bokwoon95/sq"
)

type Table struct {
	TableSchema      string
	TableName        string
	Columns          []Column
	Constraints      []Constraint
	Indexes          []Index
	Triggers         []Trigger
	VirtualTable     string
	VirtualTableArgs []string
	columnsCache     map[string]int
	constraintsCache map[string]int
	indexesCache     map[string]int
	triggersCache    map[string]int
}

func (tbl *Table) CachedColumnIndex(columnName string) (columnIndex int) {
	if columnName == "" {
		return -1
	}
	columnIndex, ok := tbl.columnsCache[columnName]
	if !ok {
		return -1
	}
	if columnIndex < 0 && columnIndex >= len(tbl.Columns) || tbl.Columns[columnIndex].ColumnName != columnName {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	return columnIndex
}

func (tbl *Table) AppendColumn(column Column) (columnIndex int) {
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	columnIndex = len(tbl.Columns) - 1
	tbl.columnsCache[column.ColumnName] = columnIndex
	return columnIndex
}

func (tbl *Table) RefreshColumnsCache() {
	for i, column := range tbl.Columns {
		if tbl.columnsCache == nil {
			tbl.columnsCache = make(map[string]int)
		}
		tbl.columnsCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintIndex(constraintName string) (constraintIndex int) {
	if constraintName == "" {
		return -1
	}
	constraintIndex, ok := tbl.constraintsCache[constraintName]
	if !ok {
		return -1
	}
	if constraintIndex < 0 || constraintIndex >= len(tbl.Constraints) || tbl.Constraints[constraintIndex].ConstraintName != constraintName {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	return constraintIndex
}

func (tbl *Table) AppendConstraint(constraint Constraint) (constraintIndex int) {
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	constraintIndex = len(tbl.Constraints) - 1
	tbl.constraintsCache[constraint.ConstraintName] = constraintIndex
	return constraintIndex
}

func (tbl *Table) RefreshConstraintCache() {
	for i, constraint := range tbl.Constraints {
		if tbl.constraintsCache == nil {
			tbl.constraintsCache = make(map[string]int)
		}
		tbl.constraintsCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexIndex(indexName string) (indexIndex int) {
	if indexName == "" {
		return -1
	}
	indexIndex, ok := tbl.indexesCache[indexName]
	if !ok {
		return -1
	}
	if indexIndex < 0 || indexIndex >= len(tbl.Indexes) || tbl.Indexes[indexIndex].IndexName != indexName {
		delete(tbl.indexesCache, indexName)
		return -1
	}
	return indexIndex
}

func (tbl *Table) AppendIndex(index Index) (indexIndex int) {
	tbl.Indexes = append(tbl.Indexes, index)
	if tbl.indexesCache == nil {
		tbl.indexesCache = make(map[string]int)
	}
	indexIndex = len(tbl.Indexes) - 1
	tbl.indexesCache[index.IndexName] = indexIndex
	return indexIndex
}

func (tbl *Table) RefreshIndexCache() {
	for i, index := range tbl.Indexes {
		if tbl.indexesCache == nil {
			tbl.indexesCache = make(map[string]int)
		}
		tbl.indexesCache[index.IndexName] = i
	}
}

func (tbl *Table) CachedTriggerIndex(triggerName string) (triggerIndex int) {
	if triggerName == "" {
		return -1
	}
	triggerIndex, ok := tbl.triggersCache[triggerName]
	if !ok {
		return -1
	}
	if triggerIndex < 0 || triggerIndex >= len(tbl.Triggers) || tbl.Triggers[triggerIndex].TriggerName != triggerName {
		delete(tbl.triggersCache, triggerName)
		return -1
	}
	return triggerIndex
}

func (tbl *Table) AppendTrigger(trigger Trigger) (triggerIndex int) {
	tbl.Triggers = append(tbl.Triggers, trigger)
	if tbl.triggersCache == nil {
		tbl.triggersCache = make(map[string]int)
	}
	triggerIndex = len(tbl.Triggers) - 1
	tbl.triggersCache[trigger.TriggerName] = triggerIndex
	return triggerIndex
}

func (tbl *Table) RefreshTriggerCache() {
	for i, trigger := range tbl.Triggers {
		if tbl.triggersCache == nil {
			tbl.triggersCache = make(map[string]int)
		}
		tbl.triggersCache[trigger.TriggerName] = i
	}
}

func (tbl *Table) LoadIndexConfig(tableSchema, tableName string, columns []string, config string) error {
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
		index = tbl.Indexes[i]
		defer func(i int) { tbl.Indexes[i] = index }(i)
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
			index.Where = modifier[1]
		case "include":
			index.Include = strings.Split(modifier[1], ",")
		default:
			return fmt.Errorf("invalid modifier 'index.%s'", modifier[0])
		}
	}
	return nil
}

func (tbl *Table) LoadConstraintConfig(constraintType, tableSchema, tableName string, columns []string, config string) error {
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

func (tbl *Table) LoadColumnConfig(dialect, columnName, columnType, config string) error {
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
