package ddl3

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
)

type IndexDiff struct {
	TableSchema    string
	TableName      string
	IndexName      string
	IndexType      string
	CreateCommand  *CreateIndexCommand
	DropCommand    *DropIndexCommand
	RenameCommand  *RenameIndexCommand
	ReplaceCommand *RenameIndexCommand
}

type CreateIndexCommand struct {
	CreateConcurrently bool
	CreateIfNotExists  bool
	Index              Index
}

func (cmd *CreateIndexCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString("CREATE")
	isFulltextOrSpatial := cmd.Index.IndexType == "FULLTEXT" || cmd.Index.IndexType == "SPATIAL"
	if dialect == sq.DialectMySQL && isFulltextOrSpatial {
		buf.WriteString(" " + cmd.Index.IndexType)
	} else if cmd.Index.IsUnique {
		buf.WriteString(" UNIQUE")
	}
	buf.WriteString(" INDEX ")
	if cmd.CreateConcurrently {
		if dialect != sq.DialectPostgres {
			return fmt.Errorf("%s does not support CREATE INDEX CONCURRENTLY", dialect)
		}
		buf.WriteString("CONCURRENTLY ")
	}
	if cmd.CreateIfNotExists && dialect != sq.DialectMySQL {
		buf.WriteString("IF NOT EXISTS ")
	}
	buf.WriteString(cmd.Index.IndexName + " ON ")
	if cmd.Index.TableSchema != "" {
		buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Index.TableSchema) + ".")
	}
	buf.WriteString(sq.QuoteIdentifier(dialect, cmd.Index.TableName))
	if cmd.Index.IndexType != "" && !isFulltextOrSpatial && !strings.EqualFold(cmd.Index.IndexType, "BTREE") {
		buf.WriteString(" USING " + cmd.Index.IndexType)
	}
	buf.WriteString(" (")
	for i, column := range cmd.Index.Columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		if column != "" {
			buf.WriteString(column)
		} else {
			buf.WriteString(cmd.Index.Exprs[i])
		}
	}
	buf.WriteString(")")
	if len(cmd.Index.Include) > 0 && dialect == sq.DialectPostgres {
		buf.WriteString(" INCLUDE (" + strings.Join(cmd.Index.Include, ", ") + ")")
	}
	if cmd.Index.Where != "" && (dialect == sq.DialectPostgres || dialect == sq.DialectSQLite) {
		buf.WriteString(" WHERE " + cmd.Index.Where)
	}
	buf.WriteString(";")
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

type RenameIndexCommand struct {
	AlterIndexIfExists bool
	TableSchema        string
	TableName          string
	IndexName          string
	RenameToName       string
}
