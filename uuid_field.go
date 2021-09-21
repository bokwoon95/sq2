package sq

import "bytes"

type UUIDField struct {
	info FieldInfo
}

func NewUUIDField(fieldName string, tableInfo TableInfo) UUIDField {
	return UUIDField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

var _ Field = (*UUIDField)(nil)

func (f UUIDField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f UUIDField) GetAlias() string { return f.info.FieldAlias }

func (f UUIDField) GetName() string { return f.info.FieldName }

func (f UUIDField) As(alias string) UUIDField {
	f.info.FieldAlias = alias
	return f
}
