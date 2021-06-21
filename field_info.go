package sq

import (
	"bytes"
	"database/sql"
)

// TODO: move this to a file of its own, together with the `GetFieldInfo(field Field) (FieldInfo, error)` function
type FieldInfoGetter interface {
	GetFieldInfo() (FieldInfo, error)
}

type FieldInfo struct {
	TableSchema string
	TableName   string
	TableAlias  string
	FieldName   string
	FieldAlias  string
	Format      string
	Values      []interface{}
	Descending  sql.NullBool
	NullsFirst  sql.NullBool
}

var _ SQLExcludeAppender = FieldInfo{}

func (f FieldInfo) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.Format != "" {
		err := BufferPrintf(dialect, buf, args, params, excludedTableQualifiers, f.Format, f.Values)
		if err != nil {
			return err
		}
	} else {
		tableQualifier := f.TableName
		if f.TableAlias != "" {
			tableQualifier = f.TableAlias
		}
		if tableQualifier != "" {
			for _, excludedTableQualifier := range excludedTableQualifiers {
				if tableQualifier == excludedTableQualifier {
					tableQualifier = ""
					break
				}
			}
		}
		if tableQualifier != "" {
			buf.WriteString(QuoteIdentifier(dialect, tableQualifier) + ".")
		}
		buf.WriteString(QuoteIdentifier(dialect, f.FieldName))
	}
	if f.Descending.Valid {
		if f.Descending.Bool {
			buf.WriteString(" DESC")
		} else {
			buf.WriteString(" ASC")
		}
	}
	if f.NullsFirst.Valid {
		if f.NullsFirst.Bool {
			buf.WriteString(" NULLS FIRST")
		} else {
			buf.WriteString(" NULLS LAST")
		}
	}
	return nil
}
