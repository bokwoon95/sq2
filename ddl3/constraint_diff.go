package ddl3

type ConstraintDiff struct {
	TableSchema    string
	TableName      string
	ConstraintName string
	ConstraintType string
	AddCommand     *AddConstraintCommand
	DropCommand    *DropConstraintCommand
	RenameCommand  *RenameConstraintCommand
	ReplaceCommand *RenameConstraintCommand
}

type AddConstraintCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	AddIfNotExists     bool
	Constraint         Constraint
	IndexSchema        string
	IndexName          string
	IsNotValid         bool
}

type DropConstraintCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	DropIfExists       bool
	ConstraintName     string
	DropCascade        bool
}

type RenameConstraintCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	ConstraintName     string
	RenameToName       string
}