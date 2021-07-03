package ddl2

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
