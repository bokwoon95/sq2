package sq

import "bytes"

type BlobField struct {
	info FieldInfo
}

var _ Field = BlobField{}

func NewBlobField(fieldName string, tableInfo TableInfo) BlobField {
	return BlobField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func (f BlobField) GetAlias() string { return f.info.FieldAlias }

func (f BlobField) GetName() string { return f.info.FieldName }

func (f BlobField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f BlobField) As(alias string) BlobField {
	f.info.FieldAlias = alias
	return f
}

func (f BlobField) Asc() BlobField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f BlobField) Desc() BlobField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f BlobField) NullsLast() BlobField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f BlobField) NullsFirst() BlobField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f BlobField) IsNull() Predicate { return IsNull(f) }

func (f BlobField) IsNotNull() Predicate { return IsNotNull(f) }

func (f BlobField) Set(val interface{}) Assignment { return Assign(f, val) }

func (f BlobField) SetBlob(val []byte) Assignment { return Assign(f, val) }
