package sq

import (
	"bytes"
)

type CustomField struct {
	info FieldInfo
}

var _ Field = CustomField{}

func NewCustomField(fieldName string, tableInfo TableInfo) CustomField {
	return CustomField{info: FieldInfo{
		TableSchema: tableInfo.TableSchema,
		TableName:   tableInfo.TableName,
		TableAlias:  tableInfo.TableAlias,
		FieldName:   fieldName,
	}}
}

func (f CustomField) GetAlias() string { return f.info.FieldAlias }

func (f CustomField) GetName() string { return f.info.FieldName }

func (f CustomField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
}

func Fieldf(format string, values ...interface{}) CustomField {
	return CustomField{info: FieldInfo{
		Format: format,
		Values: values,
	}}
}

func FieldValue(value interface{}) CustomField { return Fieldf("{}", value) }

func (f CustomField) As(alias string) CustomField {
	f.info.FieldAlias = alias
	return f
}

func (f CustomField) Asc() CustomField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f CustomField) Desc() CustomField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f CustomField) NullsLast() CustomField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f CustomField) NullsFirst() CustomField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f CustomField) IsNull() Predicate { return IsNull(f) }

func (f CustomField) IsNotNull() Predicate { return IsNotNull(f) }

func (f CustomField) In(v interface{}) Predicate { return In(f, v) }

func (f CustomField) Eq(v interface{}) Predicate { return Eq(f, v) }

func (f CustomField) Ne(v interface{}) Predicate { return Ne(f, v) }

func (f CustomField) Gt(v interface{}) Predicate { return Gt(f, v) }

func (f CustomField) Ge(v interface{}) Predicate { return Ge(f, v) }

func (f CustomField) Lt(v interface{}) Predicate { return Lt(f, v) }

func (f CustomField) Le(v interface{}) Predicate { return Le(f, v) }

type FieldLiteral string

var _ Field = FieldLiteral("")

func (f FieldLiteral) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	buf.WriteString(string(f))
	return nil
}

func (f FieldLiteral) GetAlias() string { return "" }

func (f FieldLiteral) GetName() string { return string(f) }

type Fields []Field

var _ SQLExcludeAppender = Fields{}

func (fs Fields) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	var err error
	for i, field := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		if field == nil {
			err = BufferPrintValue(dialect, buf, args, params, excludedTableQualifiers, nil, "")
			if err != nil {
				return err
			}
		} else {
			err = field.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs Fields) AppendSQLExcludeWithAlias(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	var alias string
	var err error
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		if f == nil {
			BufferPrintValue(dialect, buf, args, params, excludedTableQualifiers, nil, "")
		} else {
			err = f.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
			if err != nil {
				return err
			}
			if alias = f.GetAlias(); alias != "" {
				buf.WriteString(" AS ")
				buf.WriteString(alias)
			}
		}
	}
	return nil
}
