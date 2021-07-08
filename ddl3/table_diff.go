package ddl3

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type TableDiff struct {
	TableSchema          string
	TableName            string
	CreateCommand        CreateTableCommand
	DropCommand          DropTableCommand
	RenameCommand        RenameTableCommand
	ColumnDiffs          []ColumnDiff
	ConstraintDiffs      []ConstraintDiff
	IndexDiffs           []IndexDiff
	DataQueries          []sq.Query
	columnDiffsCache     map[string]int
	constraintDiffsCache map[string]int
	indexDiffsCache      map[string]int
}

type CreateTableCommand struct {
	Valid              bool
	CreateIfNotExists  bool
	IncludeConstraints bool
	Table              Table
	Query              sq.Query
}

var _ Command = CreateTableCommand{}

func (cmd CreateTableCommand) ToSQL(dialect string) (string, error) {
	if !cmd.Valid {
		return "", nil
	}
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	if cmd.Table.TableName == "" {
		return "", fmt.Errorf("CREATE TABLE: table has no name")
	}
	if len(cmd.Table.Columns) == 0 {
		return "", fmt.Errorf("CREATE TABLE: table %s has no columns", cmd.Table.TableName)
	}
	if cmd.Table.VirtualTable != "" {
		if dialect != sq.DialectSQLite {
			return "", fmt.Errorf("CREATE TABLE: only SQLite has VIRTUAL TABLE support (table=%s)", cmd.Table.TableName)
		}
		buf.WriteString("CREATE VIRTUAL TABLE ")
	} else {
		buf.WriteString("CREATE TABLE ")
	}
	if cmd.CreateIfNotExists {
		buf.WriteString("IF NOT EXISTS ")
	}
	if cmd.Table.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Table.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Table.TableName))
	if cmd.Table.VirtualTable != "" {
		buf.WriteString(" USING " + cmd.Table.VirtualTable)
	}
	buf.WriteString(" (")
	return buf.String(), nil
}

type DropTableCommand struct {
	Valid        bool
	DropIfExists bool
	TableSchema  string
	TableName    string
	DropCascade  bool
}

type RenameTableCommand struct {
	Valid          bool
	RenameIfExists bool
	TableSchema    string
	TableName      string
	RenameToName   string
}
