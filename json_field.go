package sq

import "bytes"

type JSONField struct {
	info FieldInfo
}

func NewJSONField(fieldName string, tableInfo TableInfo) JSONField {
	return JSONField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

var _ Field = JSONField{}

func (f JSONField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f JSONField) GetAlias() string { return f.info.FieldAlias }

func (f JSONField) GetName() string { return f.info.FieldName }

func (f JSONField) As(alias string) JSONField {
	f.info.FieldAlias = alias
	return f
}

func (f JSONField) Asc() JSONField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f JSONField) Desc() JSONField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f JSONField) NullsLast() JSONField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f JSONField) NullsFirst() JSONField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}
