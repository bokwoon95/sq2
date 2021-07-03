package ddl2

import (
	"fmt"
	"strings"
)

type Table struct {
	TableSchema      string
	TableName        string
	Columns          []Column
	Constraints      []Constraint
	Indices          []Index
	Triggers         []Object
	VirtualTable     string
	VirtualTableArgs []string
	columnsCache     map[string]int
	constraintsCache map[string]int
	indicesCache     map[string]int
	triggersCache    map[string]int
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
			index.Where = modifier[1]
		case "include":
			index.Include = strings.Split(modifier[1], ",")
		default:
			return fmt.Errorf("invalid modifier 'index.%s'", modifier[0])
		}
	}
	return nil
}

func (tbl *Table) CachedColumnIndex(columnName string) (columnIndex int) {
	if tbl == nil {
		return -1
	}
	if columnName == "" {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	columnIndex, ok := tbl.columnsCache[columnName]
	if !ok || columnIndex < 0 && columnIndex >= len(tbl.Columns) || tbl.Columns[columnIndex].ColumnName != columnName {
		delete(tbl.columnsCache, columnName)
		return -1
	}
	return columnIndex
}

func (tbl *Table) AppendColumn(column Column) (columnIndex int) {
	if tbl == nil {
		return -1
	}
	tbl.Columns = append(tbl.Columns, column)
	if tbl.columnsCache == nil {
		tbl.columnsCache = make(map[string]int)
	}
	columnIndex = len(tbl.Columns) - 1
	tbl.columnsCache[column.ColumnName] = columnIndex
	return columnIndex
}

func (tbl *Table) RefreshColumnCache() {
	if tbl == nil {
		return
	}
	for i, column := range tbl.Columns {
		if tbl.columnsCache == nil {
			tbl.columnsCache = make(map[string]int)
		}
		tbl.columnsCache[column.ColumnName] = i
	}
}

func (tbl *Table) CachedConstraintIndex(constraintName string) (constraintIndex int) {
	if tbl == nil {
		return -1
	}
	if constraintName == "" {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	constraintIndex, ok := tbl.constraintsCache[constraintName]
	if !ok || constraintIndex < 0 || constraintIndex >= len(tbl.Constraints) || tbl.Constraints[constraintIndex].ConstraintName != constraintName {
		delete(tbl.constraintsCache, constraintName)
		return -1
	}
	return constraintIndex
}

func (tbl *Table) AppendConstraint(constraint Constraint) (constraintIndex int) {
	if tbl == nil {
		return -1
	}
	tbl.Constraints = append(tbl.Constraints, constraint)
	if tbl.constraintsCache == nil {
		tbl.constraintsCache = make(map[string]int)
	}
	constraintIndex = len(tbl.Constraints) - 1
	tbl.constraintsCache[constraint.ConstraintName] = constraintIndex
	return constraintIndex
}

func (tbl *Table) RefreshConstraintCache() {
	if tbl == nil {
		return
	}
	for i, constraint := range tbl.Constraints {
		if tbl.constraintsCache == nil {
			tbl.constraintsCache = make(map[string]int)
		}
		tbl.constraintsCache[constraint.ConstraintName] = i
	}
}

func (tbl *Table) CachedIndexIndex(indexName string) (indexIndex int) {
	if tbl == nil {
		return -1
	}
	if indexName == "" {
		delete(tbl.indicesCache, indexName)
		return -1
	}
	indexIndex, ok := tbl.indicesCache[indexName]
	if !ok || indexIndex < 0 || indexIndex >= len(tbl.Indices) || tbl.Indices[indexIndex].IndexName != indexName {
		delete(tbl.indicesCache, indexName)
		return -1
	}
	return indexIndex
}

func (tbl *Table) AppendIndex(index Index) (indexIndex int) {
	if tbl == nil {
		return -1
	}
	tbl.Indices = append(tbl.Indices, index)
	if tbl.indicesCache == nil {
		tbl.indicesCache = make(map[string]int)
	}
	indexIndex = len(tbl.Indices) - 1
	tbl.indicesCache[index.IndexName] = indexIndex
	return indexIndex
}

func (tbl *Table) RefreshIndexCache() {
	if tbl == nil {
		return
	}
	for i, index := range tbl.Indices {
		if tbl.indicesCache == nil {
			tbl.indicesCache = make(map[string]int)
		}
		tbl.indicesCache[index.IndexName] = i
	}
}

func (tbl *Table) CachedTriggerIndex(triggerName string) (triggerIndex int) {
	if tbl == nil {
		return -1
	}
	if triggerName == "" {
		delete(tbl.triggersCache, triggerName)
		return -1
	}
	triggerIndex, ok := tbl.triggersCache[triggerName]
	if !ok || triggerIndex < 0 || triggerIndex >= len(tbl.Triggers) || tbl.Triggers[triggerIndex].Name != triggerName {
		delete(tbl.triggersCache, triggerName)
		return -1
	}
	return triggerIndex
}

func (tbl *Table) AppendTrigger(trigger Object) (triggerIndex int) {
	if tbl == nil {
		return -1
	}
	tbl.Triggers = append(tbl.Triggers, trigger)
	if tbl.triggersCache == nil {
		tbl.triggersCache = make(map[string]int)
	}
	triggerIndex = len(tbl.Triggers) - 1
	tbl.triggersCache[trigger.Name] = triggerIndex
	return triggerIndex
}

func (tbl *Table) RefreshTriggerCache() {
	if tbl == nil {
		return
	}
	for i, trigger := range tbl.Triggers {
		if tbl.triggersCache == nil {
			tbl.triggersCache = make(map[string]int)
		}
		tbl.triggersCache[trigger.Name] = i
	}
}
