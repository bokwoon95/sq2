package ddl3

import "github.com/bokwoon95/sq"

type ColumnDiff struct {
	TableSchema       string
	TableName         string
	ColumnName        string
	AddCommand        *AddColumnCommand
	AlterCommand      *AlterColumnCommand
	DropCommand       *DropColumnCommand
	RenameCommand     *RenameColumnCommand
	ReplaceCommand    *RenameColumnCommand
	DualWriteTriggers []TriggerDiff
	BackfillQueries   []sq.Query
}

type AddColumnCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	AddIfNotExists     bool
	Column             Column
	CheckExprs         []string
	ReferencesTable    string
	ReferencesColumn   string
}

// TODO: MySQL will be troublesome because its ALTER TABLE MODIFY|RENAME COLUMN
// is a bit of a problem child: you can string together multiple operations in
// the same command e.g. ALTER TABLE tbl DROP PRIMARY KEY, MODIFY COLUMN column
// BIGINT, column PRIMARY KEY;
// TODO: what if we handled all MySQL ALTER COLUMN using RENAME only? Don't
// support any of that multi-action crap. Is that enough?
// TODO: MySQL PRIMARY KEY should not be handled as a constraint, because the
// name is always PRIMARY (stupid silly MySQL).
type AlterColumnCommand struct {
	AlterTableIfExists bool
	AlterIfExists      bool
	Column             Column
	DropDefault        bool
	DropNotNull        bool
	DropExpr           bool
	DropIdentity       bool
	DropAutoincrement  bool
	UsingExpr          string
}

type DropColumnCommand struct {
	AlterTableIfExists bool
	DropIfExists       bool
	TableSchema        string
	TableName          string
	ColumnName         string
	DropCascade        bool
}

type RenameColumnCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	ColumnName         string
	RenameToName       string
}
