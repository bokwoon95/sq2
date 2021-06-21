package sq

import (
	"bytes"
)

type NumberField struct {
	GenericField
	Format string
	Values []interface{}
}

var _ Field = NumberField{}

func NewNumberField(fieldName string, tbl TableInfo) NumberField {
	return NumberField{GenericField: GenericField{
		TableSchema: tbl.TableSchema,
		TableName:   tbl.TableName,
		TableAlias:  tbl.TableAlias,
		FieldName:   fieldName,
	}}
}

func NumberFieldf(format string, values ...interface{}) NumberField {
	return NumberField{Format: format, Values: values}
}

func (f NumberField) As(alias string) NumberField {
	f.FieldAlias = alias
	return f
}

func (f NumberField) Asc() NumberField {
	f.Descending.Valid = true
	f.Descending.Bool = false
	return f
}

func (f NumberField) Desc() NumberField {
	f.Descending.Valid = true
	f.Descending.Bool = true
	return f
}

func (f NumberField) NullsLast() NumberField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = false
	return f
}

func (f NumberField) NullsFirst() NumberField {
	f.Nullsfirst.Valid = true
	f.Nullsfirst.Bool = true
	return f
}

func (f NumberField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.Format != "" {
		err := BufferPrintf(dialect, buf, args, params, excludedTableQualifiers, f.Format, f.Values)
		if err != nil {
			return err
		}
		f.TableSchema, f.TableName, f.TableAlias, f.FieldName, f.FieldAlias = "", "", "", "", ""
	}
	return f.GenericField.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func (f NumberField) IsNull() Predicate { return IsNull(f) }

func (f NumberField) IsNotNull() Predicate { return IsNotNull(f) }

func (f NumberField) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("{} IN {}", f, v)
	}
	return Predicatef("{} IN ({})", f, v)
}

func (f NumberField) Eq(field NumberField) Predicate { return Eq(f, field) }

func (f NumberField) Ne(field NumberField) Predicate { return Ne(f, field) }

func (f NumberField) Gt(field NumberField) Predicate { return Gt(f, field) }

func (f NumberField) Ge(field NumberField) Predicate { return Ge(f, field) }

func (f NumberField) Lt(field NumberField) Predicate { return Lt(f, field) }

func (f NumberField) Le(field NumberField) Predicate { return Le(f, field) }

func (f NumberField) EqInt(val int) Predicate { return Eq(f, val) }

func (f NumberField) NeInt(val int) Predicate { return Ne(f, val) }

func (f NumberField) GtInt(val int) Predicate { return Gt(f, val) }

func (f NumberField) GeInt(val int) Predicate { return Ge(f, val) }

func (f NumberField) LtInt(val int) Predicate { return Lt(f, val) }

func (f NumberField) LeInt(val int) Predicate { return Le(f, val) }

func (f NumberField) EqInt64(val int64) Predicate { return Eq(f, val) }

func (f NumberField) NeInt64(val int64) Predicate { return Ne(f, val) }

func (f NumberField) GtInt64(val int64) Predicate { return Gt(f, val) }

func (f NumberField) GeInt64(val int64) Predicate { return Ge(f, val) }

func (f NumberField) LtInt64(val int64) Predicate { return Lt(f, val) }

func (f NumberField) LeInt64(val int64) Predicate { return Le(f, val) }

func (f NumberField) EqFloat64(val float64) Predicate { return Eq(f, val) }

func (f NumberField) NeFloat64(val float64) Predicate { return Ne(f, val) }

func (f NumberField) GtFloat64(val float64) Predicate { return Gt(f, val) }

func (f NumberField) GeFloat64(val float64) Predicate { return Ge(f, val) }

func (f NumberField) LtFloat64(val float64) Predicate { return Lt(f, val) }

func (f NumberField) LeFloat64(val float64) Predicate { return Le(f, val) }

func (f NumberField) SetInt(val int) Assignment { return Assign(f, val) }

func (f NumberField) SetInt64(val int64) Assignment { return Assign(f, val) }

func (f NumberField) SetFloat64(val float64) Assignment { return Assign(f, val) }
