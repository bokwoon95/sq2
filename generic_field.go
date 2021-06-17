package sq

import (
	"bytes"
	"database/sql"
)

type GenericField struct {
	TableSchema string
	TableName   string
	TableAlias  string
	FieldName   string
	FieldAlias  string
	Descending  sql.NullBool
	Nullsfirst  sql.NullBool
}

var _ Field = GenericField{}

func (f GenericField) GetAlias() string { return f.FieldAlias }

func (f GenericField) GetName() string { return f.FieldName }

func (f GenericField) As(alias string) GenericField {
	f.FieldAlias = alias
	return f
}

func (f GenericField) Asc() GenericField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f GenericField) Desc() GenericField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f GenericField) NullsLast() GenericField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f GenericField) NullsFirst() GenericField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}

func (f GenericField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
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
	if f.Descending.Valid {
		if f.Descending.Bool {
			buf.WriteString(" DESC")
		} else {
			buf.WriteString(" ASC")
		}
	}
	if f.Nullsfirst.Valid {
		if f.Nullsfirst.Bool {
			buf.WriteString(" NULLS FIRST")
		} else {
			buf.WriteString(" NULLS LAST")
		}
	}
	return nil
}

func (f GenericField) IsNull() Predicate { return IsNull(f) }

func (f GenericField) IsNotNull() Predicate { return IsNotNull(f) }

func (f GenericField) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("{} IN {}", f, v)
	}
	return Predicatef("{} IN ({})", f, v)
}

func (f GenericField) Set(val interface{}) Assignment { return Assign(f, val) }
