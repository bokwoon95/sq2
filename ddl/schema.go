package ddl

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type Schema struct {
	SchemaName    string      `json:",omitempty"`
	Tables        []*Table    `json:",omitempty"`
	Views         []*View     `json:",omitempty"`
	Functions     []*Function `json:",omitempty"`
	Ignore        bool        `json:",omitempty"`
	tableCache    map[string]int
	viewCache     map[string]int
	functionCache map[string][]int
}

func (s *Schema) CachedTablePosition(tableName string) (tablePosition int) {
	if tableName == "" {
		return -1
	}
	tablePosition, ok := s.tableCache[tableName]
	if !ok {
		return -1
	}
	if tablePosition < 0 || tablePosition >= len(s.Tables) || s.Tables[tablePosition].TableName != tableName {
		delete(s.tableCache, tableName)
		return -1
	}
	return tablePosition
}

func (s *Schema) AppendTable(table *Table) (tablePosition int) {
	s.Tables = append(s.Tables, table)
	if s.tableCache == nil {
		s.tableCache = make(map[string]int)
	}
	tablePosition = len(s.Tables) - 1
	s.tableCache[table.TableName] = tablePosition
	return tablePosition
}

func (s *Schema) RefreshTableCache() {
	if s.tableCache == nil && len(s.Tables) > 0 {
		s.tableCache = make(map[string]int)
	}
	for i, table := range s.Tables {
		s.tableCache[table.TableName] = i
	}
}

func (s *Schema) CachedViewPosition(viewName string) (viewPosition int) {
	if viewName == "" {
		return -1
	}
	viewPosition, ok := s.viewCache[viewName]
	if !ok {
		return -1
	}
	if viewPosition < 0 || viewPosition >= len(s.Views) || s.Views[viewPosition].ViewName != viewName {
		delete(s.viewCache, viewName)
		return -1
	}
	return viewPosition
}

func (s *Schema) AppendView(view *View) (viewPosition int) {
	s.Views = append(s.Views, view)
	if s.viewCache == nil {
		s.viewCache = make(map[string]int)
	}
	viewPosition = len(s.Views) - 1
	s.viewCache[view.ViewName] = viewPosition
	return viewPosition
}

func (s *Schema) RefreshViewCache() {
	if s.viewCache == nil && len(s.Views) > 0 {
		s.viewCache = make(map[string]int)
	}
	for i, view := range s.Views {
		s.viewCache[view.ViewName] = i
	}
	return
}

func (s *Schema) CachedFunctionPositions(functionName string) (functionPositions []int) {
	if functionName == "" {
		return nil
	}
	functionPositions, ok := s.functionCache[functionName]
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
		s.functionCache[functionName] = functionPositions
	}
	return functionPositions
}

func (s *Schema) AppendFunction(function *Function) (functionPositions int) {
	s.Functions = append(s.Functions, function)
	if s.functionCache == nil {
		s.functionCache = make(map[string][]int)
	}
	functionPositions = len(s.Functions) - 1
	s.functionCache[function.FunctionName] = append(s.functionCache[function.FunctionName], functionPositions)
	return functionPositions
}

func (s *Schema) RefreshFunctionCache() {
	if s.functionCache == nil && len(s.Functions) > 0 {
		s.functionCache = make(map[string][]int)
	}
	for i, function := range s.Functions {
		s.functionCache[function.FunctionName] = append(s.functionCache[function.FunctionName], i)
	}
}

type CreateSchemaCommand struct {
	CreateIfNotExists bool
	SchemaName        string
}

func (cmd *CreateSchemaCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support CREATE SCHEMA")
	}
	buf.WriteString("CREATE SCHEMA ")
	if cmd.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(cmd.SchemaName + ";")
	return nil
}

type DropSchemaCommand struct {
	DropIfExists bool
	SchemaName   string
	DropCascade  bool
}

type RenameSchemaCommand struct {
	SchemaName   string
	RenameToName string
}
