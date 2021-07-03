package ddl

type Metadata struct {
	Dialect         string
	VersionString   string
	VersionNum      [2]int // MAJOR.MINOR (are PATCH versions ever significant in the case of databases?)
	GeneratedFromDB bool
	Schemas         []Schema
	schemasCache    map[string]int
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

type Schema struct {
	SchemaName     string
	Tables         []Table
	Views          []Object
	Functions      []Object
	tablesCache    map[string]int
	viewsCache     map[string]int
	functionsCache map[string]int
}

func (s *Schema) CachedTableIndex(tableName string) (tableIndex int) {
	if s == nil {
		return -1
	}
	if tableName == "" {
		delete(s.tablesCache, tableName)
		return -1
	}
	tableIndex, ok := s.tablesCache[tableName]
	if !ok || tableIndex < 0 || tableIndex >= len(s.Tables) || s.Tables[tableIndex].TableName != tableName {
		delete(s.tablesCache, tableName)
		return -1
	}
	return tableIndex
}

func (s *Schema) AppendTable(table Table) (tableIndex int) {
	if s == nil {
		return -1
	}
	s.Tables = append(s.Tables, table)
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	tableIndex = len(s.Tables) - 1
	s.tablesCache[table.TableName] = tableIndex
	return tableIndex
}

func (s *Schema) RefreshTableCache() {
	if s == nil {
		return
	}
	for i, table := range s.Tables {
		if s.tablesCache == nil {
			s.tablesCache = make(map[string]int)
		}
		s.tablesCache[table.TableName] = i
	}
}

func (s *Schema) CachedViewIndex(viewName string) (viewIndex int) {
	if s == nil {
		return -1
	}
	if viewName == "" {
		delete(s.viewsCache, viewName)
		return -1
	}
	viewIndex, ok := s.viewsCache[viewName]
	if !ok || viewIndex < 0 || viewIndex >= len(s.Views) || s.Views[viewIndex].Name != viewName {
		delete(s.viewsCache, viewName)
		return -1
	}
	return viewIndex
}

func (s *Schema) AppendView(view Object) (viewIndex int) {
	if s == nil {
		return -1
	}
	s.Views = append(s.Views, view)
	if s.viewsCache == nil {
		s.viewsCache = make(map[string]int)
	}
	viewIndex = len(s.Views) - 1
	s.viewsCache[view.Name] = viewIndex
	return viewIndex
}

func (s *Schema) RefreshViewCache() {
	if s == nil {
		return
	}
	for i, view := range s.Views {
		if s.viewsCache == nil {
			s.viewsCache = make(map[string]int)
		}
		s.viewsCache[view.Name] = i
	}
	return
}

func (s *Schema) CachedFunctionIndex(functionName string) (functionIndex int) {
	if s == nil {
		return -1
	}
	if functionName == "" {
		delete(s.functionsCache, functionName)
		return -1
	}
	functionIndex, ok := s.functionsCache[functionName]
	if !ok || functionIndex < 0 || functionIndex >= len(s.Functions) || s.Functions[functionIndex].Name != functionName {
		delete(s.functionsCache, functionName)
		return -1
	}
	return functionIndex
}

func (s *Schema) AppendFunction(function Object) (functionIndex int) {
	if s == nil {
		return -1
	}
	s.Functions = append(s.Functions, function)
	if s.functionsCache == nil {
		s.functionsCache = make(map[string]int)
	}
	functionIndex = len(s.Functions) - 1
	s.functionsCache[function.Name] = functionIndex
	return functionIndex
}

func (s *Schema) RefreshFunctionCache() {
	if s == nil {
		return
	}
	for i, function := range s.Functions {
		if s.functionsCache == nil {
			s.functionsCache = make(map[string]int)
		}
		s.functionsCache[function.Name] = i
	}
}

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
