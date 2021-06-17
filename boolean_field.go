package sq

import (
	"bytes"
)

type BooleanField struct {
	GenericField
	Negative bool
}

var _ Field = BooleanField{}

func NewBooleanField(fieldName string, tbl GenericTable) BooleanField {
	return BooleanField{GenericField: GenericField{
		TableSchema: tbl.TableSchema,
		TableName:   tbl.TableName,
		TableAlias:  tbl.TableAlias,
		FieldName:   fieldName,
	}}
}

func (f BooleanField) As(alias string) BooleanField {
	f.FieldAlias = alias
	return f
}

func (f BooleanField) Not() Predicate {
	f.Negative = !f.Negative
	return f
}

func (f BooleanField) Asc() BooleanField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f BooleanField) Desc() BooleanField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f BooleanField) NullsLast() BooleanField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f BooleanField) NullsFirst() BooleanField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}

func (f BooleanField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.Negative {
		buf.WriteString("NOT ")
	}
	return f.GenericField.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func (f BooleanField) IsNull() Predicate { return IsNull(f) }

func (f BooleanField) IsNotNull() Predicate { return IsNotNull(f) }

func (f BooleanField) Eq(field BooleanField) Predicate { return Eq(f, field) }

func (f BooleanField) Ne(field BooleanField) Predicate { return Ne(f, field) }

func (f BooleanField) EqBool(val bool) Predicate { return Eq(f, val) }

func (f BooleanField) NeBool(val bool) Predicate { return Ne(f, val) }

func (f BooleanField) SetBool(val bool) Assignment { return Assign(f, val) }
