package ddl3

import (
	"io"

	"github.com/bokwoon95/sq"
)

type CatalogDiff struct {
	SchemaDiffs      []SchemaDiff
	schemaDiffsCache map[string]int
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

type SchemaDiff struct {
	SchemaName         string
	CreateCommand      Command
	DropCommand        Command
	RenameCommand      Command
	TableDiffs         []TableDiff
	ViewDiffs          []ViewDiff
	FunctionDiffs      []FunctionDiff
	tableDiffsCache    map[string]int
	viewDiffsCache     map[string]int
	functionDiffsCache map[string][]int
}

type TableDiff struct {
	TableSchema     string
	TableName       string
	CreateCommand   Command
	DropCommand     Command
	RenameCommand   Command
	ColumnDiffs     []ColumnDiff
	ConstraintDiffs []ConstraintDiff
	IndexDiffs      []IndexDiff
	DataQueries     []sq.Query
}

type ColumnDiff struct {
	TableSchema   string
	TableName     string
	ColumnName    string
	AddCommand    Command
	AlterCommand  Command
	DropCommand   Command
	RenameCommand Command
}

type ConstraintDiff struct {
	TableSchema    string
	TableName      string
	ConstraintName string
	ConstraintType string
	AddCommand     Command
	DropCommand    Command
	RenameCommand  Command
}

type IndexDiff struct {
	TableSchema   string
	TableName     string
	IndexName     string
	IndexType     string
	CreateCommand Command
	DropCommand   Command
	RenameCommand Command
}

type TriggerDiff struct {
	TableSchema   string
	TableName     string
	TriggerName   string
	Commands      []Command
	CreateCommand Command
	DropCommand   Command
	RenameCommand Command
}

type ViewDiff struct {
	ViewSchema    string
	ViewName      string
	Commands      []Command
	CreateCommand Command
	DropCommand   Command
	RenameCommand Command
}

type FunctionDiff struct {
	FunctionSchema string
	FunctionName   string
	Commands       []Command
	CreateCommand  Command
	DropCommand    Command
	RenameCommand  Command
}
