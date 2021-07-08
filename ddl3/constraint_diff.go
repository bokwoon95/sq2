package ddl3

type ConstraintDiff struct {
	TableSchema    string
	TableName      string
	ConstraintName string
	ConstraintType string
	AddCommand     AddConstraintCommand
	DropCommand    DropConstraintCommand
	RenameCommand  RenameConstraintCommand
}

type AddConstraintCommand struct {
	Valid              bool
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
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	DropIfExists       bool
	ConstraintName     string
	DropCascade        bool
}

type RenameConstraintCommand struct {
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	ConstraintName     string
	RenameToName       string
}
