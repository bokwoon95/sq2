package sq

import (
	"bytes"
)

type BooleanField struct {
	info     FieldInfo
	Negative bool
}

var _ Field = BooleanField{}

func NewBooleanField(fieldName string, tableInfo TableInfo) BooleanField {
	return BooleanField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func (f BooleanField) GetAlias() string { return f.info.FieldAlias }

func (f BooleanField) GetName() string { return f.info.FieldName }

func (f BooleanField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	if f.Negative {
		buf.WriteString("NOT ")
	}
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f BooleanField) As(alias string) BooleanField {
	f.info.FieldAlias = alias
	return f
}

func (f BooleanField) Not() Predicate {
	f.Negative = !f.Negative
	return f
}

func (f BooleanField) Asc() BooleanField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f BooleanField) Desc() BooleanField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f BooleanField) NullsLast() BooleanField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f BooleanField) NullsFirst() BooleanField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f BooleanField) IsNull() Predicate { return IsNull(f) }

func (f BooleanField) IsNotNull() Predicate { return IsNotNull(f) }

func (f BooleanField) Eq(field BooleanField) Predicate { return Eq(f, field) }

func (f BooleanField) Ne(field BooleanField) Predicate { return Ne(f, field) }

func (f BooleanField) EqBool(val bool) Predicate { return Eq(f, val) }

func (f BooleanField) NeBool(val bool) Predicate { return Ne(f, val) }

func (f BooleanField) Set(val interface{}) Assignment { return Assign(f, val) }

func (f BooleanField) SetBool(val bool) Assignment { return Assign(f, val) }
