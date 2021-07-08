package ddl3

type ColumnDiff struct {
	TableSchema    string
	TableName      string
	ColumnName     string
	AddCommand     AddColumnCommand
	AlterCommand   AlterColumnCommand
	DropCommand    DropColumnCommand
	RenameCommand  RenameColumnCommand
	ReplaceCommand RenameColumnCommand
}

type AddColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	AddIfNotExists     bool
	Column             Column
	CheckExprs         []string
	ReferencesTable    string
	ReferencesColumn   string
}

type AlterColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	AlterIfExists      bool
	Column             Column
	DropDefault        bool
	DropNull           bool
	DropExpr           bool
	DropIdentity       bool
	UsingExpr          string
}

type DropColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	DropIfExists       bool
	TableSchema        string
	TableName          string
	ColumnName         string
	DropCascade        bool
}

type RenameColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	ColumnName         string
	RenameToName       string
}
