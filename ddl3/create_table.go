package ddl3

import (
	"bytes"
	"fmt"

	"github.com/bokwoon95/sq"
)

type CreateTableCommand struct {
	Dialect       string
	DoIfNotExists bool
	Table         Table
	Query         sq.Query
}

var _ Command = CreateTableCommand{}

func (cmd CreateTableCommand) GetType() string { return CREATE_TABLE }

func (cmd CreateTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
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
	if cmd.DoIfNotExists {
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
	return nil
}
