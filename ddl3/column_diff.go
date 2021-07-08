package ddl3

type ColumnDiff struct {
	TableSchema   string
	TableName     string
	ColumnName    string
	AddCommand    AddColumnCommand
	AlterCommand  AlterColumnCommand
	DropCommand   DropColumnCommand
	RenameCommand RenameColumnCommand
}

type AddColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	DoIfNotExists      bool
	Column             Column
	CheckExprs         []string
	ReferencesTable    string
	ReferencesColumn   string
}

type AlterColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	DoIfExists         bool
	Column             Column
	DropDefault        bool
	DropNull           bool
	DropExpr           bool
	DropIdentity       bool
	Using              string
}

type DropColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	DoIfExists         bool
	TableSchema        string
	TableName          string
	ColumnName         string
	DoCascade          bool
}

type RenameColumnCommand struct {
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	ColumnName         string
	RenameToName       string
}
