package ddl

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/bokwoon95/sq"
)

type Column struct {
	TableSchema              string `json:",omitempty"`
	TableName                string `json:",omitempty"`
	TableAlias               string `json:",omitempty"`
	ColumnName               string `json:",omitempty"`
	ColumnAlias              string `json:",omitempty"`
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

var _ sq.Field = Column{}

func (c Column) GetName() string { return c.ColumnName }

func (c Column) GetAlias() string { return c.ColumnAlias }

func (c Column) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	tableQualifier := c.TableAlias
	if tableQualifier == "" {
		tableQualifier = c.TableName
	}
	if tableQualifier != "" {
		i := sort.SearchStrings(excludedTableQualifiers, tableQualifier)
		if i < len(excludedTableQualifiers) && excludedTableQualifiers[i] == tableQualifier {
			tableQualifier = ""
		}
	}
	if tableQualifier != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, tableQualifier) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, c.ColumnName))
	return nil
}

type AddColumnCommand struct {
	AddIfNotExists   bool
	Column           Column
	CheckExprs       []string
	ReferencesTable  string
	ReferencesColumn string
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

type AlterColumnCommand struct {
	Column            Column
	DropDefault       bool
	DropNotNull       bool
	DropExpr          bool
	DropExprIfExists  bool
	DropIdentity      bool
	DropAutoincrement bool
	UsingExpr         string
}

type DropColumnCommand struct {
	DropIfExists bool
	ColumnName   string
	DropCascade  bool
}

type RenameColumnCommand struct {
	ColumnName   string
	RenameToName string
}
