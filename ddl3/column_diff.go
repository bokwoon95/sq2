package ddl3

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

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

func (cmd *AddColumnCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("ALTER TABLE ")
	if cmd.Column.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Column.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Column.TableName) + " ADD COLUMN " + sq.QuoteIdentifier(dialect, cmd.Column.ColumnName))
	err := writeColumn(dialect, buf, cmd.Column)
	if err != nil {
		return fmt.Errorf("ADD COLUMN: %w", err)
	}
	buf.WriteString(";")
	return nil
}

func writeColumn(dialect string, buf *bytes.Buffer, column Column) error {
	buf.WriteString(sq.QuoteIdentifier(dialect, column.ColumnName))
	if column.ColumnType != "" {
		buf.WriteString(" " + column.ColumnType)
	}
	if column.IsNotNull {
		buf.WriteString(" NOT NULL")
	}
	if column.ColumnDefault != "" && !column.Autoincrement && column.Identity == "" && column.GeneratedExpr == "" {
		buf.WriteString(" DEFAULT " + column.ColumnDefault)
	}
	if column.IsPrimaryKey && dialect == sq.DialectSQLite {
		// only SQLite primary key is defined inline, others are defined as separate constraints
		buf.WriteString(" PRIMARY KEY")
	}
	if column.Autoincrement && dialect != sq.DialectMySQL && dialect != sq.DialectSQLite {
		return fmt.Errorf("%s does not support autoincrement columns", dialect)
	}
	if column.Identity != "" && (dialect == sq.DialectMySQL || dialect == sq.DialectSQLite) {
		return fmt.Errorf("%s does not support identity columns", dialect)
	}
	if column.Autoincrement {
		switch dialect {
		case sq.DialectMySQL:
			buf.WriteString(" AUTO_INCREMENT")
		case sq.DialectSQLite:
			buf.WriteString(" AUTOINCREMENT")
		}
	} else if column.Identity != "" {
		buf.WriteString(" GENERATED " + column.Identity)
	} else if column.GeneratedExpr != "" {
		buf.WriteString(" GENERATED ALWAYS AS (" + column.GeneratedExpr + ")")
		if column.GeneratedExprStored {
			buf.WriteString(" STORED")
		} else {
			if dialect == sq.DialectPostgres {
				return fmt.Errorf("Postgres does not support VIRTUAL generated columns")
			}
			buf.WriteString(" VIRTUAL")
		}
	}
	if column.OnUpdateCurrentTimestamp {
		if dialect != sq.DialectMySQL {
			return fmt.Errorf("%s does not support ON UPDATE CURRENT_TIMESTAMP", dialect)
		}
		buf.WriteString(" ON UPDATE CURRENT_TIMESTAMP")
	}
	if column.CollationName != "" {
		switch dialect {
		case sq.DialectPostgres:
			buf.WriteString(` COLLATE "` + sq.EscapeQuote(column.CollationName, '"') + `"`) // postgres collation names need double quotes (idk why)
		default:
			buf.WriteString(" COLLATE " + column.CollationName)
		}
	}
	return nil
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
