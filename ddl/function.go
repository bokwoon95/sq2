package ddl

import (
	"bytes"
	"fmt"
	"strings"
)

type Function struct {
	FunctionSchema string `json:",omitempty"`
	FunctionName   string `json:",omitempty"`
	ContainsTable  bool   `json:",omitempty"`
	SQL            string `json:",omitempty"`
}

func getFunctionInfo(dialect, sql string) (functionSchema, functionName string, err error) {
	const (
		PRE_FUNCTION = iota
		FUNCTION
	)
	word, rest := "", sql
	state := PRE_FUNCTION
LOOP:
	for rest != "" {
		switch state {
		case PRE_FUNCTION:
			word, rest = popWord(dialect, rest)
			if strings.EqualFold(word, "FUNCTION") {
				state = FUNCTION
			}
			continue
		case FUNCTION:
			functionName, rest = popWord(dialect, rest)
			if i := strings.IndexByte(functionName, '.'); i >= 0 {
				functionSchema, functionName = functionName[:i], functionName[i+1:]
			}
			if i := strings.IndexByte(functionName, '('); i >= 0 {
				functionName = functionName[:i]
			}
			break LOOP
		}
	}
	if functionName == "" {
		return functionSchema, functionName, fmt.Errorf("could not find function name, did you write the function correctly?")
	}
	return functionSchema, functionName, nil
}

type FunctionDiff struct {
	FunctionSchema string
	FunctionName   string
	CreateCommand  *CreateFunctionCommand
	DropCommand    *DropFunctionCommand
	RenameCommand  *RenameFunctionCommand
	ReplaceCommand *RenameFunctionCommand
}

type CreateFunctionCommand struct {
	Function Function
}

func (cmd *CreateFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(cmd.Function.SQL)
	return nil
}

type DropFunctionCommand struct {
	DropIfExists   bool
	FunctionSchema string
	FunctionName   string
	DropCascade    bool
}

type RenameFunctionCommand struct {
	FunctionSchema string
	FunctionName   string
	RenameToName   string
}
