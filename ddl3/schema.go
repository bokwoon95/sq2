package ddl3

type Schema struct {
	SchemaName     string
	Tables         []Table
	Views          []View
	Functions      []Function
	tablesCache    map[string]int
	viewsCache     map[string]int
	functionsCache map[string]int
}

func (s *Schema) CachedTableIndex(tableName string) (tableIndex int) {
	if tableName == "" {
		return -1
	}
	tableIndex, ok := s.tablesCache[tableName]
	if !ok {
		return -1
	}
	if tableIndex < 0 || tableIndex >= len(s.Tables) || s.Tables[tableIndex].TableName != tableName {
		delete(s.tablesCache, tableName)
		return -1
	}
	return tableIndex
}

func (s *Schema) AppendTable(table Table) (tableIndex int) {
	s.Tables = append(s.Tables, table)
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	tableIndex = len(s.Tables) - 1
	s.tablesCache[table.TableName] = tableIndex
	return tableIndex
}

func (s *Schema) RefreshTableCache() {
	for i, table := range s.Tables {
		if s.tablesCache == nil {
			s.tablesCache = make(map[string]int)
		}
		s.tablesCache[table.TableName] = i
	}
}

func (s *Schema) CachedViewIndex(viewName string) (viewIndex int) {
	if viewName == "" {
		return -1
	}
	viewIndex, ok := s.viewsCache[viewName]
	if !ok {
		return -1
	}
	if viewIndex < 0 || viewIndex >= len(s.Views) || s.Views[viewIndex].ViewName != viewName {
		delete(s.viewsCache, viewName)
		return -1
	}
	return viewIndex
}

func (s *Schema) AppendView(view View) (viewIndex int) {
	s.Views = append(s.Views, view)
	if s.viewsCache == nil {
		s.viewsCache = make(map[string]int)
	}
	viewIndex = len(s.Views) - 1
	s.viewsCache[view.ViewName] = viewIndex
	return viewIndex
}

func (s *Schema) RefreshViewCache() {
	for i, view := range s.Views {
		if s.viewsCache == nil {
			s.viewsCache = make(map[string]int)
		}
		s.viewsCache[view.ViewName] = i
	}
	return
}

func (s *Schema) CachedFunctionIndex(functionName string) (functionIndex int) {
	if functionName == "" {
		return -1
	}
	functionIndex, ok := s.functionsCache[functionName]
	if !ok {
		return -1
	}
	if functionIndex < 0 || functionIndex >= len(s.Functions) || s.Functions[functionIndex].FunctionName != functionName {
		delete(s.functionsCache, functionName)
		return -1
	}
	return functionIndex
}

func (s *Schema) AppendFunction(function Function) (functionIndex int) {
	s.Functions = append(s.Functions, function)
	if s.functionsCache == nil {
		s.functionsCache = make(map[string]int)
	}
	functionIndex = len(s.Functions) - 1
	s.functionsCache[function.FunctionName] = functionIndex
	return functionIndex
}

func (s *Schema) RefreshFunctionCache() {
	for i, function := range s.Functions {
		if s.functionsCache == nil {
			s.functionsCache = make(map[string]int)
		}
		s.functionsCache[function.FunctionName] = i
	}
}
