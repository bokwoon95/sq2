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

type CommandType int

const (
	DROP_SCHEMA     = "DROP SCHEMA"
	DROP_TABLE      = "DROP TABLE"
	DROP_COLUMN     = "ALTER TABLE DROP COLUMN"
	DROP_CONSTRAINT = "ALTER TABLE DROP CONSTRAINT"
	DROP_INDEX      = "DROP INDEX"
	DROP_VIEW       = "DROP VIEW"
	DROP_FUNCTION   = "DROP FUNCTION"
	DROP_TRIGGER    = "DROP TRIGGER"

	CREATE_SCHEMA   = "CREATE SCHEMA"
	CREATE_TABLE    = "CREATE TABLE"
	ADD_COLUMN      = "ALTER TABLE ADD COLUMN"
	ADD_CONSTRAINT  = "ALTER TABLE ADD CONSTRAINT"
	CREATE_INDEX    = "CREATE INDEX"
	CREATE_VIEW     = "CREATE VIEW"
	CREATE_FUNCTION = "CREATE FUNCTION"
	CREATE_TRIGGER  = "CREATE TRIGGER"

	ALTER_COLUMN = "ALTER TABLE ALTER COLUMN"
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
