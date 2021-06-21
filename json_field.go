package sq

type JSONField struct {
	GenericField
}

func NewJSONField(fieldName string, tbl TableInfo) JSONField {
	return JSONField{GenericField: GenericField{
		TableSchema: tbl.TableSchema,
		TableName:   tbl.TableName,
		TableAlias:  tbl.TableAlias,
		FieldName:   fieldName,
	}}
}

var _ Field = JSONField{}

func (f JSONField) As(alias string) JSONField {
	f.FieldAlias = alias
	return f
}

func (f JSONField) Asc() JSONField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f JSONField) Desc() JSONField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f JSONField) NullsLast() JSONField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f JSONField) NullsFirst() JSONField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}
