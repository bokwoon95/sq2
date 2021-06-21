package sq

import "bytes"

type StringField struct {
	GenericField
	Format string
	Values []interface{}
}

var _ Field = StringField{}

func NewStringField(fieldName string, tableInfo TableInfo) StringField {
	return StringField{GenericField: GenericField{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func StringFieldf(format string, values ...interface{}) StringField {
	return StringField{Format: format, Values: values}
}

func (f StringField) As(alias string) StringField {
	f.FieldAlias = alias
	return f
}

func (f StringField) Asc() StringField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f StringField) Desc() StringField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f StringField) NullsLast() StringField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f StringField) NullsFirst() StringField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}

func (f StringField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.Format != "" {
		err := BufferPrintf(dialect, buf, args, params, excludedTableQualifiers, f.Format, f.Values)
		if err != nil {
			return err
		}
		f.TableSchema, f.TableName, f.TableAlias, f.FieldName, f.FieldAlias = "", "", "", "", ""
	}
	return f.GenericField.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func (f StringField) IsNull() Predicate { return IsNull(f) }

func (f StringField) IsNotNull() Predicate { return IsNotNull(f) }

func (f StringField) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("{} IN {}", f, v)
	}
	return Predicatef("{} IN ({})", f, v)
}

func (f StringField) Eq(field StringField) Predicate { return Eq(f, field) }

func (f StringField) Ne(field StringField) Predicate { return Ne(f, field) }

func (f StringField) EqString(val string) Predicate { return Eq(f, val) }

func (f StringField) NeString(val string) Predicate { return Ne(f, val) }

func (f StringField) SetString(val string) Assignment { return Assign(f, val) }
