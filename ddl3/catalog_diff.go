package ddl3

import (
	"io"
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

type CommandType int

// TODO: For (sqlite,postgres,mysql) does renaming a table:
// - break triggers?
// - break views?
const (
	// Schema
	CREATE_SCHEMA CommandType = 1 << iota
	RENAME_SCHEMA
	DROP_SCHEMA
	// Table
	CREATE_TABLE
	RENAME_TABLE
	DROP_TABLE
	// Column
	ADD_COLUMN
	ALTER_COLUMN
	RENAME_COLUMN
	DROP_COLUMN
	// Constraint
	ADD_CONSTRAINT
	RENAME_CONSTRAINT
	DROP_CONSTRAINT
	// Index
	CREATE_INDEX
	RENAME_INDEX
	DROP_INDEX
	// Function
	CREATE_FUNCTION
	RENAME_FUNCTION
	DROP_FUNCTION
	// View
	CREATE_VIEW
	RENAME_VIEW
	DROP_VIEW
	// Trigger
	CREATE_TRIGGER
	RENAME_TRIGGER
	DROP_TRIGGER
	// Table DML
	TABLE_DML
)

/*
Creation granularity is table-centric
- CreateSchema
- CreateTable
	- AddColumn
	- AlterColumn
	- AddConstraint
	- CreateIndex
- CreateView
- CreateFunction
- CreateTrigger
- DML
- RenameX
- DropX
*/
