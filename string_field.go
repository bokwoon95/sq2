package sq

import "bytes"

type StringField struct {
	info FieldInfo
}

var _ Field = StringField{}

func (f StringField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f StringField) GetAlias() string { return f.info.FieldAlias }

func (f StringField) GetName() string { return f.info.FieldName }

func NewStringField(fieldName string, tableInfo TableInfo) StringField {
	return StringField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func StringFieldf(format string, values ...interface{}) StringField {
	return StringField{info: FieldInfo{
		Formats: [][2]string{{"default", format}},
		Values:  values,
	}}
}

func (f StringField) As(alias string) StringField {
	f.info.FieldAlias = alias
	return f
}

func (f StringField) Collate(collation string) StringField {
	f.info.Collation = collation
	return f
}

func (f StringField) Asc() StringField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f StringField) Desc() StringField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f StringField) NullsLast() StringField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f StringField) NullsFirst() StringField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f StringField) IsNull() Predicate { return IsNull(f) }

func (f StringField) IsNotNull() Predicate { return IsNotNull(f) }

func (f StringField) In(v interface{}) Predicate { return In(f, v) }

func (f StringField) Eq(field StringField) Predicate { return Eq(f, field) }

func (f StringField) Ne(field StringField) Predicate { return Ne(f, field) }

func (f StringField) EqString(val string) Predicate { return Eq(f, val) }

func (f StringField) NeString(val string) Predicate { return Ne(f, val) }

func (f StringField) LikeString(val string) Predicate { return Predicatef("{} LIKE {}", f, val) }

func (f StringField) ILikeString(val string) Predicate { return Predicatef("{} ILIKE {}", f, val) }

func (f StringField) Set(val interface{}) Assignment { return Assign(f, val) }

func (f StringField) SetString(val string) Assignment { return Assign(f, val) }
