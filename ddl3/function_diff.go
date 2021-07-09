package ddl3

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
