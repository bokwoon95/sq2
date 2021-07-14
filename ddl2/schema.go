package ddl2

type Schema struct {
	SchemaName     string     `json:",omitempty"`
	Tables         []Table    `json:",omitempty"`
	Views          []View     `json:",omitempty"`
	Functions      []Function `json:",omitempty"`
	tablesCache    map[string]int
	viewsCache     map[string]int
	functionsCache map[string][]int
}

func (s *Schema) CachedTablePosition(tableName string) (tablePosition int) {
	if tableName == "" {
		return -1
	}
	tablePosition, ok := s.tablesCache[tableName]
	if !ok {
		return -1
	}
	if tablePosition < 0 || tablePosition >= len(s.Tables) || s.Tables[tablePosition].TableName != tableName {
		delete(s.tablesCache, tableName)
		return -1
	}
	return tablePosition
}

func (s *Schema) AppendTable(table Table) (tablePosition int) {
	s.Tables = append(s.Tables, table)
	if s.tablesCache == nil {
		s.tablesCache = make(map[string]int)
	}
	tablePosition = len(s.Tables) - 1
	s.tablesCache[table.TableName] = tablePosition
	return tablePosition
}

func (s *Schema) RefreshTableCache() {
	for i, table := range s.Tables {
		if s.tablesCache == nil {
			s.tablesCache = make(map[string]int)
		}
		s.tablesCache[table.TableName] = i
	}
}

func (s *Schema) CachedViewPosition(viewName string) (viewPosition int) {
	if viewName == "" {
		return -1
	}
	viewPosition, ok := s.viewsCache[viewName]
	if !ok {
		return -1
	}
	if viewPosition < 0 || viewPosition >= len(s.Views) || s.Views[viewPosition].ViewName != viewName {
		delete(s.viewsCache, viewName)
		return -1
	}
	return viewPosition
}

func (s *Schema) AppendView(view View) (viewPosition int) {
	s.Views = append(s.Views, view)
	if s.viewsCache == nil {
		s.viewsCache = make(map[string]int)
	}
	viewPosition = len(s.Views) - 1
	s.viewsCache[view.ViewName] = viewPosition
	return viewPosition
}

func (s *Schema) RefreshViewsCache() {
	for i, view := range s.Views {
		if s.viewsCache == nil {
			s.viewsCache = make(map[string]int)
		}
		s.viewsCache[view.ViewName] = i
	}
	return
}

func (s *Schema) CachedFunctionPositions(functionName string) (functionPositions []int) {
	if functionName == "" {
		return nil
	}
	functionPositions, ok := s.functionsCache[functionName]
	if !ok {
		return nil
	}
	var n int
	var hasInvalidPosition bool
	for _, i := range functionPositions {
		if i < 0 || i >= len(s.Functions) || s.Functions[i].FunctionName != functionName {
			hasInvalidPosition = true
			continue
		}
		functionPositions[n] = i
		n++
	}
	if hasInvalidPosition {
		functionPositions = functionPositions[:n]
		s.functionsCache[functionName] = functionPositions
	}
	return functionPositions
}

func (s *Schema) AppendFunction(function Function) (functionPositions int) {
	s.Functions = append(s.Functions, function)
	if s.functionsCache == nil {
		s.functionsCache = make(map[string][]int)
	}
	functionPositions = len(s.Functions) - 1
	s.functionsCache[function.FunctionName] = append(s.functionsCache[function.FunctionName], functionPositions)
	return functionPositions
}

func (s *Schema) RefreshFunctionsCache() {
	for i, function := range s.Functions {
		if s.functionsCache == nil {
			s.functionsCache = make(map[string][]int)
		}
		s.functionsCache[function.FunctionName] = append(s.functionsCache[function.FunctionName], i)
	}
}
