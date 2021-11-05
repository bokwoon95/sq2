package sq

import "bytes"

type BinaryField struct {
	info FieldInfo
}

var _ Field = BinaryField{}

func NewBlobField(fieldName string, tableInfo TableInfo) BinaryField {
	return BinaryField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func (f BinaryField) GetAlias() string { return f.info.FieldAlias }

func (f BinaryField) GetName() string { return f.info.FieldName }

func (f BinaryField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f BinaryField) As(alias string) BinaryField {
	f.info.FieldAlias = alias
	return f
}

func (f BinaryField) Asc() BinaryField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f BinaryField) Desc() BinaryField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f BinaryField) NullsLast() BinaryField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f BinaryField) NullsFirst() BinaryField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f BinaryField) IsNull() Predicate { return IsNull(f) }

func (f BinaryField) IsNotNull() Predicate { return IsNotNull(f) }

func (f BinaryField) Set(val interface{}) Assignment { return Assign(f, val) }

func (f BinaryField) SetBlob(val []byte) Assignment { return Assign(f, val) }
