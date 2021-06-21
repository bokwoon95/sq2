package sq

import (
	"bytes"
	"time"
)

type TimeField struct {
	GenericField
	Format string
	Values []interface{}
}

var _ Field = TimeField{}

func NewTimeField(fieldName string, tableInfo TableInfo) TimeField {
	return TimeField{GenericField: GenericField{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func TimeFieldf(format string, values ...interface{}) TimeField {
	return TimeField{Format: format, Values: values}
}

func (f TimeField) As(alias string) TimeField {
	f.FieldAlias = alias
	return f
}

func (f TimeField) Asc() TimeField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f TimeField) Desc() TimeField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f TimeField) NullsLast() TimeField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f TimeField) NullsFirst() TimeField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}

func (f TimeField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.Format != "" {
		err := BufferPrintf(dialect, buf, args, params, excludedTableQualifiers, f.Format, f.Values)
		if err != nil {
			return err
		}
		f.TableSchema, f.TableName, f.TableAlias, f.FieldName, f.FieldAlias = "", "", "", "", ""
	}
	return f.GenericField.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func (f TimeField) IsNull() Predicate { return IsNull(f) }

func (f TimeField) IsNotNull() Predicate { return IsNotNull(f) }

func (f TimeField) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("{} IN {}", f, v)
	}
	return Predicatef("{} IN ({})", f, v)
}

func (f TimeField) Eq(field TimeField) Predicate { return Eq(f, field) }

func (f TimeField) Ne(field TimeField) Predicate { return Ne(f, field) }

func (f TimeField) Gt(field TimeField) Predicate { return Gt(f, field) }

func (f TimeField) Ge(field TimeField) Predicate { return Ge(f, field) }

func (f TimeField) Lt(field TimeField) Predicate { return Lt(f, field) }

func (f TimeField) Le(field TimeField) Predicate { return Le(f, field) }

func (f TimeField) EqTime(val time.Time) Predicate { return Eq(f, val) }

func (f TimeField) NeTime(val time.Time) Predicate { return Ne(f, val) }

func (f TimeField) GtTime(val time.Time) Predicate { return Gt(f, val) }

func (f TimeField) GeTime(val time.Time) Predicate { return Ge(f, val) }

func (f TimeField) LtTime(val time.Time) Predicate { return Lt(f, val) }

func (f TimeField) LeTime(val time.Time) Predicate { return Le(f, val) }

func (f TimeField) SetTime(val time.Time) Assignment { return Assign(f, val) }
