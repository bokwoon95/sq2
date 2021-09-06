package sq

import (
	"bytes"
	"database/sql"
	"fmt"
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
	Formats     [][2]string
	Values      []interface{}
	Collation   string
	Descending  sql.NullBool
	NullsFirst  sql.NullBool
	Err         error
}

var _ Field = FieldInfo{}

func (f FieldInfo) GetAlias() string { return f.FieldAlias }

func (f FieldInfo) GetName() string { return f.FieldName }

func (f FieldInfo) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	if f.Err != nil {
		return f.Err
	}
	if len(f.Formats) > 0 {
		var dialectFormat, defaultFormat sql.NullString
		for _, tuple := range f.Formats {
			switch tuple[0] {
			case "default":
				defaultFormat.Valid = true
				defaultFormat.String = tuple[1]
			case dialect:
				dialectFormat.Valid = true
				dialectFormat.String = tuple[1]
			}
		}
		if !defaultFormat.Valid {
			return fmt.Errorf("CustomField formats %+v has no default format", f.Formats)
		}
		format := dialectFormat.String
		if !dialectFormat.Valid {
			format = defaultFormat.String
		}
		err := BufferPrintf(dialect, buf, args, params, env, excludedTableQualifiers, format, f.Values)
		if err != nil {
			return err
		}
	} else {
		tableQualifier := f.TableAlias
		if tableQualifier == "" {
			tableQualifier = f.TableName
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
	if f.Collation != "" {
		buf.WriteString(" COLLATE ")
		if dialect == DialectPostgres {
			buf.WriteString(`"` + EscapeQuote(f.Collation, '"') + `"`)
		} else {
			buf.WriteString(QuoteIdentifier(dialect, f.Collation))
		}
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
