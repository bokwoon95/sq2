package sq

type BlobField struct {
	GenericField
}

var _ Field = BlobField{}

func NewBlobField(fieldName string, tbl TableInfo) BlobField {
	return BlobField{GenericField: GenericField{
		TableSchema: tbl.TableSchema,
		TableName:   tbl.TableName,
		TableAlias:  tbl.TableAlias,
		FieldName:   fieldName,
	}}
}

func (f BlobField) As(alias string) BlobField {
	f.FieldAlias = alias
	return f
}

func (f BlobField) Asc() BlobField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f BlobField) Desc() BlobField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f BlobField) NullsLast() BlobField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f BlobField) NullsFirst() BlobField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}

func (f BlobField) SetBlob(val []byte) Assignment { return Assign(f, val) }
