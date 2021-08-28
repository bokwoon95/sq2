package sq

import (
	"bytes"
)

type NumberField struct {
	info FieldInfo
}

var _ Field = NumberField{}

func (f NumberField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f NumberField) GetAlias() string { return f.info.FieldAlias }

func (f NumberField) GetName() string { return f.info.FieldName }

func NewNumberField(fieldName string, tableInfo TableInfo) NumberField {
	return NumberField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func NumberFieldf(format string, values ...interface{}) NumberField {
	return NumberField{info: FieldInfo{
		Formats: [][2]string{{"default", format}},
		Values:  values,
	}}
}

func (f NumberField) As(alias string) NumberField {
	f.info.FieldAlias = alias
	return f
}

func (f NumberField) Asc() NumberField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f NumberField) Desc() NumberField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f NumberField) NullsLast() NumberField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f NumberField) NullsFirst() NumberField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f NumberField) IsNull() Predicate { return IsNull(f) }

func (f NumberField) IsNotNull() Predicate { return IsNotNull(f) }

func (f NumberField) In(v interface{}) Predicate { return In(f, v) }

func (f NumberField) Eq(field NumberField) Predicate { return Eq(f, field) }

func (f NumberField) Ne(field NumberField) Predicate { return Ne(f, field) }

func (f NumberField) Gt(field NumberField) Predicate { return Gt(f, field) }

func (f NumberField) Ge(field NumberField) Predicate { return Ge(f, field) }

func (f NumberField) Lt(field NumberField) Predicate { return Lt(f, field) }

func (f NumberField) Le(field NumberField) Predicate { return Le(f, field) }

func (f NumberField) EqInt(val int) Predicate { return Eq(f, val) }

func (f NumberField) NeInt(val int) Predicate { return Ne(f, val) }

func (f NumberField) GtInt(val int) Predicate { return Gt(f, val) }

func (f NumberField) GeInt(val int) Predicate { return Ge(f, val) }

func (f NumberField) LtInt(val int) Predicate { return Lt(f, val) }

func (f NumberField) LeInt(val int) Predicate { return Le(f, val) }

func (f NumberField) EqInt64(val int64) Predicate { return Eq(f, val) }

func (f NumberField) NeInt64(val int64) Predicate { return Ne(f, val) }

func (f NumberField) GtInt64(val int64) Predicate { return Gt(f, val) }

func (f NumberField) GeInt64(val int64) Predicate { return Ge(f, val) }

func (f NumberField) LtInt64(val int64) Predicate { return Lt(f, val) }

func (f NumberField) LeInt64(val int64) Predicate { return Le(f, val) }

func (f NumberField) EqFloat64(val float64) Predicate { return Eq(f, val) }

func (f NumberField) NeFloat64(val float64) Predicate { return Ne(f, val) }

func (f NumberField) GtFloat64(val float64) Predicate { return Gt(f, val) }

func (f NumberField) GeFloat64(val float64) Predicate { return Ge(f, val) }

func (f NumberField) LtFloat64(val float64) Predicate { return Lt(f, val) }

func (f NumberField) LeFloat64(val float64) Predicate { return Le(f, val) }

func (f NumberField) Set(val interface{}) Assignment { return Assign(f, val) }

func (f NumberField) SetInt(val int) Assignment { return Assign(f, val) }

func (f NumberField) SetInt64(val int64) Assignment { return Assign(f, val) }

func (f NumberField) SetFloat64(val float64) Assignment { return Assign(f, val) }
