package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Constraint struct {
	TableSchema         string   `json:",omitempty"`
	TableName           string   `json:",omitempty"`
	ConstraintName      string   `json:",omitempty"`
	ConstraintType      string   `json:",omitempty"`
	Columns             []string `json:",omitempty"`
	Exprs               []string `json:",omitempty"`
	ReferencesSchema    string   `json:",omitempty"`
	ReferencesTable     string   `json:",omitempty"`
	ReferencesColumns   []string `json:",omitempty"`
	OnUpdate            string   `json:",omitempty"`
	OnDelete            string   `json:",omitempty"`
	MatchOption         string   `json:",omitempty"`
	CheckExpr           string   `json:",omitempty"`
	Operators           []string `json:",omitempty"`
	IndexType           string   `json:",omitempty"`
	Where               string   `json:",omitempty"`
	IsDeferrable        bool     `json:",omitempty"`
	IsInitiallyDeferred bool     `json:",omitempty"`
}

type AddConstraintCommand struct {
	Constraint Constraint
	IndexName  string
	IsNotValid bool
}

func (cmd *AddConstraintCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect == sq.DialectSQLite {
		return fmt.Errorf("sqlite does not allow constraints to be added after table creation")
	}
	buf.WriteString("ADD CONSTRAINT " + sq.QuoteIdentifier(dialect, cmd.Constraint.ConstraintName))
	if cmd.IndexName != "" {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not allow the creation of constraints using an index", dialect)
		}
		if cmd.Constraint.ConstraintType != PRIMARY_KEY && cmd.Constraint.ConstraintType != UNIQUE {
			return fmt.Errorf("postgres only allows PIRMARY KEY and UNIQUE constraints to be added using an index")
		}
		buf.WriteString(" " + cmd.Constraint.ConstraintType + " USING INDEX " + sq.QuoteIdentifier(dialect, cmd.IndexName))
		if cmd.Constraint.IsDeferrable {
			buf.WriteString(" DEFERRABLE")
			if cmd.Constraint.IsInitiallyDeferred {
				buf.WriteString(" INITIALLY DEFERRED")
			} else {
				buf.WriteString(" INITIALLY IMMEDIATE")
			}
		}
	} else {
		err := writeConstraintDefinition(dialect, buf, cmd.Constraint)
		if err != nil {
			return err
		}
	}
	if cmd.IsNotValid {
		switch dialect {
		case sq.DialectPostgres:
			if cmd.Constraint.ConstraintType != CHECK && cmd.Constraint.ConstraintType != FOREIGN_KEY {
				return fmt.Errorf("postgres %s constraints cannot be NOT VALID", cmd.Constraint.ConstraintType)
			}
			buf.WriteString(" NOT VALID")
		case sq.DialectMySQL:
			if cmd.Constraint.ConstraintType != CHECK {
				return fmt.Errorf("mysql %s constraints cannot be NOT ENFORCED", cmd.Constraint.ConstraintType)
			}
			buf.WriteString(" NOT ENFORCED")
		default:
			return fmt.Errorf("%s does not allow NOT VALID constraints", dialect)
		}
	}
	return nil
}

func writeConstraintDefinition(dialect string, buf *bytes.Buffer, constraint Constraint) error {
	switch constraint.ConstraintType {
	case CHECK:
		buf.WriteString(" CHECK (" + constraint.CheckExpr + ")")
	case FOREIGN_KEY:
		buf.WriteString(" FOREIGN KEY (" + strings.Join(constraint.Columns, ", ") + ") REFERENCES ")
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
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support EXCLUDE constraints", dialect)
		}
		if constraint.IndexType != "" {
			buf.WriteString(" EXCLUDE USING " + constraint.IndexType)
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
		buf.WriteString(" " + constraint.ConstraintType + " (" + strings.Join(constraint.Columns, ", ") + ")")
	}
	if constraint.IsDeferrable {
		switch dialect {
		case sq.DialectPostgres:
			if constraint.ConstraintType == CHECK {
				return fmt.Errorf("postgres CHECK constraints are not deferrable")
			}
		case sq.DialectSQLite:
			if constraint.ConstraintType != FOREIGN_KEY {
				return fmt.Errorf("sqlite %s constraints are not deferrable", constraint.ConstraintType)
			}
		default:
			return fmt.Errorf("%s does not support deferrable constraints", dialect)
		}
		buf.WriteString(" DEFERRABLE")
		if constraint.IsInitiallyDeferred {
			buf.WriteString(" INITIALLY DEFERRED")
		} else {
			buf.WriteString(" INITIALLY IMMEDIATE")
		}
	}
	return nil
}

type AlterConstraintCommand struct {
	ConstraintName      string
	AlterDeferrable     bool
	IsDeferrable        bool
	IsInitiallyDeferred bool
}

type DropConstraintCommand struct {
	DropIfExists   bool
	ConstraintName string
	DropCascade    bool
}

type RenameConstraintCommand struct {
	ConstraintName string
	RenameToName   string
}
