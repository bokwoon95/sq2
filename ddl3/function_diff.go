package ddl3

type FunctionDiff struct {
	FunctionSchema string
	FunctionName   string
	CreateCommand  CreateFunctionCommand
	DropCommand    DropFunctionCommand
	RenameCommand  RenameFunctionCommand
	ReplaceCommand RenameFunctionCommand
}

type CreateFunctionCommand struct {
	Valid    bool
	Function Function
}

type DropFunctionCommand struct {
	Valid          bool
	DropIfExists   bool
	FunctionSchema string
	FunctionName   string
	DropCascade    bool
}

type RenameFunctionCommand struct {
	Valid          bool
	FunctionSchema string
	FunctionName   string
	RenameToName   string
}
