package ddl3

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/bokwoon95/sq"
)

type Table struct {
	TableSchema      string
	TableName        string
	TableAlias       string
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

var _ sq.Table = Table{}

func (tbl Table) GetAlias() string { return tbl.TableAlias }

func (tbl Table) GetName() string { return tbl.TableName }

func (tbl Table) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if tbl.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, tbl.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, tbl.TableName))
	return nil
}

func (tbl *Table) CachedColumnPositions(columnName string) (columnPosition int) {
	if columnName == "" {
		return -1
	}
	columnPosition, ok := tbl.columnsCache[columnName]
	if !ok {
		return -1
	}
	if columnPosition < 0 && columnPosition >= len(tbl.Columns) || tbl.Columns[columnPosition].ColumnName != columnName {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	return columnPosition
}

func (tbl *Table) AppendColumn(column Column) (columnPosition int) {
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	columnPosition = len(tbl.Columns) - 1
	tbl.columnsCache[column.ColumnName] = columnPosition
	return columnPosition
}

func (tbl *Table) RefreshColumnsCache() {
	for i, column := range tbl.Columns {
		if tbl.columnsCache == nil {
			tbl.columnsCache = make(map[string]int)
		}
		tbl.columnsCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintPosition(constraintName string) (constraintPosition int) {
	if constraintName == "" {
		return -1
	}
	constraintPosition, ok := tbl.constraintsCache[constraintName]
	if !ok {
		return -1
	}
	if constraintPosition < 0 || constraintPosition >= len(tbl.Constraints) || tbl.Constraints[constraintPosition].ConstraintName != constraintName {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	return constraintPosition
}

func (tbl *Table) AppendConstraint(constraint Constraint) (constraintPosition int) {
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	constraintPosition = len(tbl.Constraints) - 1
	tbl.constraintsCache[constraint.ConstraintName] = constraintPosition
	return constraintPosition
}

func (tbl *Table) RefreshConstraintsCache() {
	for i, constraint := range tbl.Constraints {
		if tbl.constraintsCache == nil {
			tbl.constraintsCache = make(map[string]int)
		}
		tbl.constraintsCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexPosition(indexName string) (indexPosition int) {
	if indexName == "" {
		return -1
	}
	indexPosition, ok := tbl.indexesCache[indexName]
	if !ok {
		return -1
	}
	if indexPosition < 0 || indexPosition >= len(tbl.Indexes) || tbl.Indexes[indexPosition].IndexName != indexName {
		delete(tbl.indexesCache, indexName)
		return -1
	}
	return indexPosition
}

func (tbl *Table) AppendIndex(index Index) (indexPosition int) {
	tbl.Indexes = append(tbl.Indexes, index)
	if tbl.indexesCache == nil {
		tbl.indexesCache = make(map[string]int)
	}
	indexPosition = len(tbl.Indexes) - 1
	tbl.indexesCache[index.IndexName] = indexPosition
	return indexPosition
}

func (tbl *Table) RefreshIndexesCache() {
	for i, index := range tbl.Indexes {
		if tbl.indexesCache == nil {
			tbl.indexesCache = make(map[string]int)
		}
		tbl.indexesCache[index.IndexName] = i
	}
}

func (tbl *Table) CachedTriggerPosition(triggerName string) (triggerPosition int) {
	if triggerName == "" {
		return -1
	}
	triggerPosition, ok := tbl.triggersCache[triggerName]
	if !ok {
		return -1
	}
	if triggerPosition < 0 || triggerPosition >= len(tbl.Triggers) || tbl.Triggers[triggerPosition].TriggerName != triggerName {
		delete(tbl.triggersCache, triggerName)
		return -1
	}
	return triggerPosition
}

func (tbl *Table) AppendTrigger(trigger Trigger) (triggerPosition int) {
	tbl.Triggers = append(tbl.Triggers, trigger)
	if tbl.triggersCache == nil {
		tbl.triggersCache = make(map[string]int)
	}
	triggerPosition = len(tbl.Triggers) - 1
	tbl.triggersCache[trigger.TriggerName] = triggerPosition
	return triggerPosition
}

func (tbl *Table) RefreshTriggersCache() {
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
	if i := tbl.CachedIndexPosition(indexName); i >= 0 {
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
	if i := tbl.CachedConstraintPosition(constraintName); i >= 0 {
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
	if i := tbl.CachedColumnPositions(columnName); i >= 0 {
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
