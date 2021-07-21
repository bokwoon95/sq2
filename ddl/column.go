package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Column struct {
	TableSchema              string `json:",omitempty"`
	TableName                string `json:",omitempty"`
	ColumnName               string `json:",omitempty"`
	ColumnType               string `json:",omitempty"`
	Precision                int    `json:",omitempty"`
	Scale                    int    `json:",omitempty"`
	Identity                 string `json:",omitempty"`
	Autoincrement            bool   `json:",omitempty"`
	IsNotNull                bool   `json:",omitempty"`
	IsUnique                 bool   `json:",omitempty"`
	IsPrimaryKey             bool   `json:",omitempty"`
	OnUpdateCurrentTimestamp bool   `json:",omitempty"`
	GeneratedExpr            string `json:",omitempty"`
	GeneratedExprStored      bool   `json:",omitempty"`
	CollationName            string `json:",omitempty"`
	ColumnDefault            string `json:",omitempty"`
	Ignore                   bool   `json:",omitempty"`
}

type AddColumnCommand struct {
	AddIfNotExists   bool
	Column           Column
	CheckExprs       []string
	ReferencesTable  string
	ReferencesColumn string
}

func (cmd *AddColumnCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("ADD COLUMN ")
	if cmd.AddIfNotExists {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support ADD COLUMN IF NOT EXISTS", dialect)
		}
		buf.WriteString("IF NOT EXISTS ")
	}
	if dialect == sq.DialectSQLite {
		if cmd.Column.IsPrimaryKey {
			return fmt.Errorf("sqlite does not allow adding a PRIMARY KEY column after table creation")
		}
		if cmd.Column.IsUnique {
			return fmt.Errorf("sqlite does not allow adding a UNIQUE column after table creation")
		}
		if cmd.Column.IsNotNull {
			if cmd.Column.ColumnDefault == "" || strings.EqualFold(cmd.Column.ColumnDefault, "NULL") {
				return fmt.Errorf("sqlite does not allow adding a NOT NULL column without a non-null DEFAULT value")
			}
			if cmd.Column.ColumnDefault[0] == '(' && cmd.Column.ColumnDefault[len(cmd.Column.ColumnDefault)-1] == ')' {
				return fmt.Errorf("sqlite does not allow adding a NOT NULL column with an expression as the DEFAULT value")
			}
		}
		if (cmd.ReferencesTable != "" || cmd.ReferencesColumn != "") && cmd.Column.ColumnDefault != "" && !strings.EqualFold(cmd.Column.ColumnDefault, "NULL") {
			return fmt.Errorf("sqlite does not allow adding a FOREIGN KEY column without a NULL DEFAULT value")
		}
		if cmd.Column.GeneratedExpr != "" && cmd.Column.GeneratedExprStored {
			return fmt.Errorf("sqlite does not allow adding GENERATED STORED columns after table creation (use GENERATED VIRTUAL)")
		}
	}
	err := writeColumnDefinition(dialect, buf, cmd.Column)
	if err != nil {
		return fmt.Errorf("ADD COLUMN: %w", err)
	}
	if dialect == sq.DialectSQLite {
		for _, checkExpr := range cmd.CheckExprs {
			buf.WriteString(" CHECK (" + checkExpr + ")")
		}
		if cmd.ReferencesTable != "" && cmd.ReferencesColumn != "" {
			buf.WriteString(" REFERENCES " +
				sq.QuoteIdentifier(dialect, cmd.ReferencesTable) +
				" (" + sq.QuoteIdentifier(dialect, cmd.ReferencesColumn) + ")",
			)
		}
	}
	return nil
}

func writeColumnDefinition(dialect string, buf *bytes.Buffer, column Column) error {
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
		// only SQLite primary key is defined inline with column, other
		// dialects will define primary key constraints separately
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

type AlterColumnCommand struct {
	Column               Column
	DropDefault          bool
	DropNotNull          bool
	DropExpr             bool
	DropExprIfExists     bool
	DropIdentity         bool
	DropIdentityIfExists bool
	DropAutoincrement    bool
	UsingExpr            string
}

func (cmd *AlterColumnCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	switch dialect {
	case sq.DialectSQLite:
		return fmt.Errorf("sqlite does not support altering columns after table creation")
	case sq.DialectMySQL:
		buf.WriteString("MODIFY COLUMN ")
		err := writeColumnDefinition(dialect, buf, cmd.Column)
		if err != nil {
			return fmt.Errorf("MODIFY COLUMN: %w", err)
		}
	case sq.DialectPostgres:
		var written bool
		// alterColumnName abstracts away the boilerplate of writing "ALTER COLUMN
		// $COLUMN_NAME" every time. It also prepends each new ALTER COLUMN with a
		// newline character (except for the first ALTER COLUMN because the initial
		// newline will be handled by the parent ALTER TABLE command).
		alterColumnName := func() {
			if !written {
				written = true
			} else {
				buf.WriteString("\n    ,")
			}
			buf.WriteString("ALTER COLUMN " + sq.QuoteIdentifier(dialect, cmd.Column.ColumnName))
		}
		if cmd.Column.ColumnType != "" {
			alterColumnName()
			buf.WriteString(" SET DATA TYPE " + cmd.Column.ColumnType)
			if cmd.Column.CollationName != "" {
				buf.WriteString(` COLLATE "` + sq.EscapeQuote(cmd.Column.CollationName, '"') + `"`)
			}
			if cmd.UsingExpr != "" {
				buf.WriteString(" USING " + cmd.UsingExpr)
			}
		}
		if cmd.DropNotNull {
			alterColumnName()
			buf.WriteString(" DROP NOT NULL")
		} else if cmd.Column.IsNotNull {
			alterColumnName()
			buf.WriteString(" SET NOT NULL")
		}
		if cmd.DropDefault {
			alterColumnName()
			buf.WriteString(" DROP DEFAULT")
		} else if cmd.Column.ColumnDefault != "" {
			alterColumnName()
			buf.WriteString(" SET DEFAULT " + cmd.Column.ColumnDefault)
		}
		if cmd.DropIdentity {
			alterColumnName()
			buf.WriteString(" DROP IDENTITY")
			if cmd.DropIdentityIfExists {
				buf.WriteString(" IF EXISTS")
			}
		} else if cmd.Column.Identity != "" {
			alterColumnName()
			buf.WriteString(" ADD GENERATED " + cmd.Column.Identity)
		}
		if cmd.DropExpr {
			alterColumnName()
			buf.WriteString(" DROP EXPRESSION")
			if cmd.DropExprIfExists {
				buf.WriteString(" IF EXISTS")
			}
		}
	default:
		return fmt.Errorf("unrecognized dialect: %s", dialect)
	}
	return nil
}

type DropColumnCommand struct {
	DropIfExists bool
	ColumnName   string
	DropCascade  bool
}

func (cmd *DropColumnCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("DROP COLUMN ")
	if cmd.DropIfExists {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support DROP COLUMN IF EXISTS", dialect)
		}
		buf.WriteString("IF EXISTS ")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.ColumnName))
	if cmd.DropCascade {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support DROP COLUMN ... CASCADE", dialect)
		}
		buf.WriteString(" CASCADE")
	}
	return nil
}

type RenameColumnCommand struct {
	ColumnName   string
	RenameToName string
}

func (cmd *RenameColumnCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("RENAME COLUMN " + sq.QuoteIdentifier(dialect, cmd.ColumnName) + " TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName))
	return nil
}
