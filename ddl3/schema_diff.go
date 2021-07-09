package ddl3

type SchemaDiff struct {
	SchemaName         string
	CreateCommand      *CreateSchemaCommand
	DropCommand        *DropSchemaCommand
	RenameCommand      *RenameSchemaCommand
	TableDiffs         []TableDiff
	ViewDiffs          []ViewDiff
	FunctionDiffs      []FunctionDiff
	tableDiffsCache    map[string]int
	viewDiffsCache     map[string]int
	functionDiffsCache map[string][]int
}

type CreateSchemaCommand struct {
	CreateIfNotExists bool
	SchemaName        string
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
