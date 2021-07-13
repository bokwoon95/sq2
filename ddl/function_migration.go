package ddl

import "bytes"

type FunctionMigration struct {
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

var _ Command = &CreateFunctionCommand{}

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
