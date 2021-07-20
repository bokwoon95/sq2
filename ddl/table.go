package ddl

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
