package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type CatalogDiff struct {
	SchemaDiffs      []SchemaDiff
	schemaDiffsCache map[string]int // 8 bytes
}

func DiffCatalog(gotCatalog, wantCatalog Catalog) (CatalogDiff, error) {
	var set CatalogDiff
	return set, nil
}

func (set CatalogDiff) Commands() Commands {
	return Commands{}
}

func (set CatalogDiff) WriteOut(w io.Writer) error {
	return nil
}

/*
TODO: implement the Rename* commands
- RenameSchemaCommand
- RenameTableCommand
- RenameColumnCommand
- RenameConstraintCommand // may not be possible
- RenameIndexCommand // may not be possible
- RenameTriggerCommand // may not be possible
- RenameViewCommand // may not be possible
- RenameFunctionCommand // may not be possible
*/

type SchemaDiff struct {
	SchemaName         string              // 16 bytes
	CreateCommand      CreateSchemaCommand // 50 bytes
	DropCommand        DropSchemaCommand   // 43 bytes
	RenameCommand      Command
	TableDiffs         []TableDiff
	ViewDiffs          []ViewDiff
	FunctionDiffs      []FunctionDiff
	tableDiffsCache    map[string]int   // 8 bytes
	viewDiffsCache     map[string]int   // 8 bytes
	functionDiffsCache map[string][]int // 8 bytes
}

type TableDiff struct {
	TableSchema          string
	TableName            string
	CreateCommand        CreateTableCommand
	DropCommand          DropTableCommand
	RenameCommand        Command
	ColumnDiffs          []ColumnDiff
	ConstraintDiffs      []ConstraintDiff
	IndexDiffs           []IndexDiff
	DataQueries          []sq.Query
	columnDiffsCache     map[string]int
	constraintDiffsCache map[string]int
	indexDiffsCache      map[string]int
}

type ColumnDiff struct {
	TableSchema   string
	TableName     string
	ColumnName    string
	AddCommand    AddColumnCommand
	AlterCommand  AlterColumnCommand
	DropCommand   DropColumnCommand
	RenameCommand Command
}

type ConstraintDiff struct {
	TableSchema    string
	TableName      string
	ConstraintName string
	ConstraintType string
	AddCommand     AddConstraintCommand
	DropCommand    DropConstraintCommand
	RenameCommand  Command
}

type IndexDiff struct {
	TableSchema   string
	TableName     string
	IndexName     string
	IndexType     string
	CreateCommand CreateIndexCommand
	DropCommand   DropIndexCommand
	RenameCommand Command
}

type TriggerDiff struct {
	TableSchema   string
	TableName     string
	TriggerName   string
	Commands      []Command
	CreateCommand CreateTriggerCommand
	DropCommand   DropTriggerCommand
	RenameCommand Command
}

type ViewDiff struct {
	ViewSchema    string
	ViewName      string
	Commands      []Command
	CreateCommand CreateViewCommand
	DropCommand   DropViewCommand
	RenameCommand Command
}

type FunctionDiff struct {
	FunctionSchema string
	FunctionName   string
	CreateCommand  CreateFunctionCommand
	DropCommand    DropFunctionCommand
	RenameCommand  Command
}
