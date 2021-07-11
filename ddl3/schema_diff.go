package ddl3

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

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

func (cmd *CreateSchemaCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support CREATE SCHEMA")
	}
	buf.WriteString("CREATE SCHEMA ")
	if cmd.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(cmd.SchemaName + ";")
	return nil
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
