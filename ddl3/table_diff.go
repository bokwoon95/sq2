package ddl3

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type TableDiff struct {
	TableSchema          string
	TableName            string
	CreateCommand        *CreateTableCommand
	DropCommand          *DropTableCommand
	RenameCommand        *RenameTableCommand
	ReplaceCommand       *RenameTableCommand
	ColumnDiffs          []ColumnDiff
	ConstraintDiffs      []ConstraintDiff
	IndexDiffs           []IndexDiff
	TriggerDiffs         []TriggerDiff
	DualWriteTriggers    []TriggerDiff
	BackfillQueries      []sq.Query
	columnDiffsCache     map[string]int
	constraintDiffsCache map[string]int
	indexDiffsCache      map[string]int
}

type CreateTableCommand struct {
	CreateIfNotExists  bool
	IncludeConstraints bool
	Table              Table
	// Query              sq.Query
}

var _ Command = &CreateTableCommand{}

func (cmd *CreateTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if cmd == nil {
		return nil
	}
	if cmd.Table.TableName == "" {
		return fmt.Errorf("CREATE TABLE: table has no name")
	}
	if len(cmd.Table.Columns) == 0 {
		return fmt.Errorf("CREATE TABLE: table %s has no columns", cmd.Table.TableName)
	}
	if cmd.Table.VirtualTable != "" {
		if dialect != sq.DialectSQLite {
			return fmt.Errorf("CREATE TABLE: only SQLite has VIRTUAL TABLE support (table=%s)", cmd.Table.TableName)
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
	var columnWritten bool
	for i, column := range cmd.Table.Columns {
		if column.Ignore {
			continue
		}
		if !columnWritten {
			columnWritten = true
			buf.WriteString("\n    ")
		} else {
			buf.WriteString("\n    ,")
		}
		if strings.EqualFold(cmd.Table.VirtualTable, "fts5") {
			column = Column{ColumnName: column.ColumnName}
		}
		err := writeColumn(dialect, buf, column)
		if err != nil {
			return fmt.Errorf("column #%d: %w", i+1, err)
		}
	}
	if len(cmd.Table.VirtualTableArgs) > 0 && cmd.Table.VirtualTable == "" {
		return fmt.Errorf("virtual table arguments present without a virtual table module")
	}
	if cmd.Table.VirtualTable != "" && dialect == sq.DialectSQLite && len(cmd.Table.VirtualTableArgs) > 0 {
		for _, arg := range cmd.Table.VirtualTableArgs {
			if !columnWritten {
				columnWritten = true
				buf.WriteString("\n    ")
			} else {
				buf.WriteString("\n    ,")
			}
			buf.WriteString(arg)
		}
	}
	if cmd.IncludeConstraints {
		var newlineWritten bool
		for i, constraint := range cmd.Table.Constraints {
			if dialect == sq.DialectSQLite && constraint.ConstraintType == PRIMARY_KEY && len(constraint.Columns) == 1 {
				// SQLite PRIMARY KEY is always be defined inline with the column,
				// so we don't have to do it here.
				continue
			}
			if dialect != sq.DialectSQLite && constraint.ConstraintType == FOREIGN_KEY {
				// FOREIGN KEYs are always defined after all tables have been
				// created, to avoid referencing tables that have yet to be
				// created. SQLite is the exception because constraints cannot
				// be defined outside of CREATE TABLE. However, SQLite foreign
				// keys can be created even if the referencing tables do not
				// yet exist, so it's not an issue.
				// http://sqlite.1065341.n5.nabble.com/Circular-foreign-keys-td14977.html
				continue
			}
			if !newlineWritten {
				buf.WriteString("\n")
				newlineWritten = true
			}
			buf.WriteString("\n    ,CONSTRAINT ")
			err := writeConstraint(dialect, buf, constraint)
			if err != nil {
				return fmt.Errorf("constraint #%d: %w", i+1, err)
			}
		}
	}
	buf.WriteString("\n);")
	return nil
}

type DropTableCommand struct {
	DropIfExists bool
	TableSchema  string
	TableName    string
	DropCascade  bool
}

type RenameTableCommand struct {
	RenameIfExists bool
	TableSchema    string
	TableName      string
	RenameToName   string
}
