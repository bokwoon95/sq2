package ddl

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type Index struct {
	TableSchema    string   `json:",omitempty"`
	TableName      string   `json:",omitempty"`
	IndexName      string   `json:",omitempty"`
	IndexType      string   `json:",omitempty"`
	IsUnique       bool     `json:",omitempty"`
	Columns        []string `json:",omitempty"`
	Exprs          []string `json:",omitempty"`
	IncludeColumns []string `json:",omitempty"`
	Predicate      string   `json:",omitempty"`
	SQL            string   `json:",omitempty"`
	Ignore         bool     `json:",omitempty"`
}

type CreateIndexCommand struct {
	CreateConcurrently bool
	CreateIfNotExists  bool
	Index              Index
}

func (cmd *CreateIndexCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if dialect != sq.DialectMySQL {
		buf.WriteString("CREATE ")
	}
	isFulltextOrSpatial := cmd.Index.IndexType == "FULLTEXT" || cmd.Index.IndexType == "SPATIAL"
	if dialect == sq.DialectMySQL && isFulltextOrSpatial {
		buf.WriteString(cmd.Index.IndexType + " ")
	} else if cmd.Index.IsUnique {
		buf.WriteString("UNIQUE ")
	}
	buf.WriteString("INDEX ")
	if cmd.CreateConcurrently {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support CREATE INDEX CONCURRENTLY", dialect)
		}
		buf.WriteString("CONCURRENTLY ")
	}
	if cmd.CreateIfNotExists {
		if dialect == sq.DialectMySQL {
			return fmt.Errorf("mysql index creation does not support IF NOT EXISTS")
		}
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Index.IndexName))
	if dialect != sq.DialectMySQL {
		buf.WriteString(" ON ")
		if cmd.Index.TableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Index.TableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Index.TableName))
	}
	if cmd.Index.IndexType != "" && !isFulltextOrSpatial && !strings.EqualFold(cmd.Index.IndexType, "BTREE") {
		if dialect != sq.DialectPostgres && dialect != sq.DialectMySQL {
			return fmt.Errorf("%s does not support index types", dialect)
		}
		buf.WriteString(" USING " + cmd.Index.IndexType)
	}
	buf.WriteString(" (")
	for i, column := range cmd.Index.Columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		if column != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, column))
		} else {
			buf.WriteString(cmd.Index.Exprs[i])
		}
	}
	buf.WriteString(")")
	if len(cmd.Index.IncludeColumns) > 0 {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support INDEX ... INCLUDE", dialect)
		}
		buf.WriteString(" INCLUDE (")
		for i, column := range cmd.Index.IncludeColumns {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(sq.QuoteIdentifier(dialect, column))
		}
		buf.WriteString(")")
	}
	if cmd.Index.Predicate != "" {
		if dialect != sq.DialectPostgres && dialect != sq.DialectSQLite {
			return fmt.Errorf("%s does not support INDEX ... WHERE", dialect)
		}
		buf.WriteString(" WHERE " + cmd.Index.Predicate)
	}
	return nil
}

type DropIndexCommand struct {
	DropConcurrently bool
	DropIfExists     bool
	TableSchema      string
	TableName        string
	IndexName        string
	DropCascade      bool
}

func (cmd DropIndexCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("DROP INDEX ")
	if cmd.DropConcurrently {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support DROP INDEX CONCURRENTLY", dialect)
		}
		buf.WriteString("CONCURRENTLY ")
	}
	if cmd.DropIfExists {
		if dialect != sq.DialectPostgres && dialect != sq.DialectSQLite {
			return fmt.Errorf("%s does not support DROP INDEX IF EXISTS", dialect)
		}
		buf.WriteString("IF EXISTS ")
	}
	if cmd.TableSchema != "" && dialect != sq.DialectMySQL {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.IndexName))
	if cmd.DropCascade {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support DROP INDEX CASCADE", dialect)
		}
		buf.WriteString(" CASCADE")
	}
	return nil
}

type RenameIndexCommand struct {
	AlterIfExists bool
	TableSchema   string
	TableName     string
	IndexName     string
	RenameToName  string
}

func (cmd RenameIndexCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	switch dialect {
	case sq.DialectSQLite:
		return fmt.Errorf("sqlite does not support renaming indexes")
	case sq.DialectMySQL:
		buf.WriteString("RENAME INDEX " + sq.QuoteIdentifier(dialect, cmd.IndexName) + " TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName))
	default:
		buf.WriteString("ALTER INDEX ")
		if cmd.AlterIfExists {
			buf.WriteString("IF EXISTS ")
		}
		if cmd.TableSchema != "" {
			buf.WriteString(sq.QuoteIdentifier(dialect, cmd.TableSchema) + ".")
		}
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.IndexName) + " RENAME TO " + sq.QuoteIdentifier(dialect, cmd.RenameToName))
	}
	return nil
}
