package ddl3

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type ConstraintDiff struct {
	TableSchema    string
	TableName      string
	ConstraintName string
	ConstraintType string
	AddCommand     *AddConstraintCommand
	DropCommand    *DropConstraintCommand
	RenameCommand  *RenameConstraintCommand
	ReplaceCommand *RenameConstraintCommand
}

type AddConstraintCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	AddIfNotExists     bool
	Constraint         Constraint
	IndexSchema        string
	IndexName          string
	IsNotValid         bool
}

func (cmd *AddConstraintCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("SQLite does not allow the creating of constraints separately")
	}
	buf.WriteString("ALTER TABLE ")
	if cmd.Constraint.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Constraint.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Constraint.TableName) + " ADD CONSTRAINT ")
	err := writeConstraint(dialect, buf, cmd.Constraint)
	if err != nil {
		return err
	}
	buf.WriteString(";")
	return nil
}

func writeConstraint(dialect string, buf *bytes.Buffer, constraint Constraint) error {
	buf.WriteString(constraint.ConstraintName + " " + constraint.ConstraintType)
	switch constraint.ConstraintType {
	case CHECK:
		buf.WriteString(" (" + constraint.CheckExpr + ")")
	case FOREIGN_KEY:
		buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ") REFERENCES ")
		if constraint.ReferencesSchema != "" {
			buf.WriteString(constraint.ReferencesSchema + ".")
		}
		buf.WriteString(constraint.ReferencesTable)
		if len(constraint.ReferencesColumns) > 0 {
			buf.WriteString(" (" + strings.Join(constraint.ReferencesColumns, ", ") + ")")
		}
		if constraint.MatchOption != "" {
			buf.WriteString(" " + constraint.MatchOption)
		}
		if constraint.OnUpdate != "" {
			buf.WriteString(" ON UPDATE " + constraint.OnUpdate)
		}
		if constraint.OnDelete != "" {
			buf.WriteString(" ON DELETE " + constraint.OnDelete)
		}
	case EXCLUDE:
		if constraint.IndexType != "" {
			buf.WriteString(" USING " + constraint.IndexType)
		}
		buf.WriteString(" (")
		for i := range constraint.Columns {
			if i > 0 {
				buf.WriteString(", ")
			}
			if column := constraint.Columns[i]; column != "" {
				buf.WriteString(column)
			} else if expr := constraint.Exprs[i]; expr != "" {
				buf.WriteString(expr)
			} else {
				return fmt.Errorf("column #%d: no column name or expression provided", i+1)
			}
			buf.WriteString(" WITH ")
			if operator := constraint.Operators[i]; operator != "" {
				buf.WriteString(operator)
			} else {
				return fmt.Errorf("column #%d: no exclusion operator provided", i+1)
			}
		}
		buf.WriteString(")")
		if constraint.Where != "" {
			buf.WriteString(" WHERE (" + constraint.Where + ")")
		}
	default:
		buf.WriteString(" (" + strings.Join(constraint.Columns, ", ") + ")")
	}
	var deferSupported bool
	if (dialect == sq.DialectPostgres && constraint.ConstraintType != CHECK) ||
		(dialect == sq.DialectSQLite && constraint.ConstraintType == FOREIGN_KEY) {
		deferSupported = true
	}
	if deferSupported && constraint.IsDeferrable {
		buf.WriteString(" DEFERRABLE")
		if constraint.IsInitiallyDeferred {
			buf.WriteString(" INITIALLY DEFERRED")
		} else {
			buf.WriteString(" INITIALLY IMMEDIATE")
		}
	}
	return nil
}

type DropConstraintCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	DropIfExists       bool
	ConstraintName     string
	DropCascade        bool
}

type RenameConstraintCommand struct {
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	ConstraintName     string
	RenameToName       string
}
