package sq

import (
	"bytes"
	"time"
)

type TimeField struct {
	info FieldInfo
}

var _ Field = TimeField{}

func (f TimeField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f TimeField) GetAlias() string { return f.info.FieldAlias }

func (f TimeField) GetName() string { return f.info.FieldName }

func NewTimeField(fieldName string, tableInfo TableInfo) TimeField {
	return TimeField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func TimeFieldf(format string, values ...interface{}) TimeField {
	return TimeField{info: FieldInfo{
		Formats: [][2]string{{"default", format}},
		Values:  values,
	}}
}

func (f TimeField) As(alias string) TimeField {
	f.info.FieldAlias = alias
	return f
}

func (f TimeField) Asc() TimeField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f TimeField) Desc() TimeField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f TimeField) NullsLast() TimeField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f TimeField) NullsFirst() TimeField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f TimeField) IsNull() Predicate { return IsNull(f) }

func (f TimeField) IsNotNull() Predicate { return IsNotNull(f) }

func (f TimeField) In(v interface{}) Predicate { return In(f, v) }

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

func (f TimeField) Set(val interface{}) Assignment { return Assign(f, val) }

func (f TimeField) SetTime(val time.Time) Assignment { return Assign(f, val) }
