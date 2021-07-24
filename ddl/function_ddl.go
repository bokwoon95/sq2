package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type CreateFunctionCommand struct {
	Function Function
}

func (cmd CreateFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support functions")
	}
	buf.WriteString(cmd.Function.SQL)
	return nil
}

type DropFunctionCommand struct {
	DropIfExists bool
	Function     Function
	DropCascade  bool
}

func (cmd DropFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not support functions")
	}
	buf.WriteString("DROP FUNCTION ")
	if cmd.DropIfExists {
		buf.WriteString("IF EXISTS ")
	}
	if cmd.Function.FunctionSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionName))
	if dialect == sq.DialectPostgres {
		buf.WriteString("(" + strings.Join(cmd.Function.ArgTypes, ", ") + ")")
	}
	if cmd.DropCascade {
		buf.WriteString(" CASCADE")
	}
	buf.WriteString(";")
	return nil
}

type RenameFunctionCommand struct {
	Function     Function
	RenameToName string
}

func (cmd RenameFunctionCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite || dialect == sq.DialectMySQL {
		return fmt.Errorf("%s does not support renaming functions", dialect)
	}
	buf.WriteString("ALTER FUNCTION ")
	if cmd.Function.FunctionSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Function.FunctionName))
	if dialect == sq.DialectPostgres {
		buf.WriteString("(" + strings.Join(cmd.Function.ArgTypes, ", ") + ")")
	}
	buf.WriteString(" RENAME TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName) + ";")
	return nil
}
