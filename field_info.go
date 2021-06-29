package sq

import (
	"bytes"
	"database/sql"
	"sort"
)

// TODO: move this to a file of its own, together with the `GetFieldInfo(field Field) (FieldInfo, error)` function
// first try to type assert to a field.(FieldInfo) directly
// then see if it is one of the builtin types: BlobField | BooleanField | GenericField | JSONField | NumberField | StringField | TimeField
// then see if it implements FieldInfoGetter
// else return error
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
	StickyErr   error
}

var _ Field = FieldInfo{}

func (f FieldInfo) GetAlias() string { return f.FieldAlias }

func (f FieldInfo) GetName() string { return f.FieldName }

func (f FieldInfo) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.StickyErr != nil {
		return f.StickyErr
	}
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
			i := sort.SearchStrings(excludedTableQualifiers, tableQualifier)
			if i < len(excludedTableQualifiers) && excludedTableQualifiers[i] == tableQualifier {
				tableQualifier = ""
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
