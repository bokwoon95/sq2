package sq

import (
	"bytes"
	"fmt"
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

func (f CustomField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func Fieldf(format string, values ...interface{}) CustomField {
	return CustomField{info: FieldInfo{
		Formats: [][2]string{{"default", format}},
		Values:  values,
	}}
}

func FieldfDialect(formats map[string]string, values ...interface{}) CustomField {
	customField := CustomField{info: FieldInfo{
		Formats: make([][2]string, 0, len(formats)),
		Values:  values,
	}}
	for dialect, format := range formats {
		customField.info.Formats = append(customField.info.Formats, [2]string{dialect, format})
	}
	return customField
}

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

type FieldLiteral struct {
	literal string
	alias   string
}

var _ Field = FieldLiteral{}

func (f FieldLiteral) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	buf.WriteString(f.literal)
	return nil
}

func (f FieldLiteral) GetAlias() string { return f.alias }

func (f FieldLiteral) GetName() string { return f.literal }

func Literal(literal string) FieldLiteral {
	return FieldLiteral{literal: literal}
}

func (f FieldLiteral) As(alias string) FieldLiteral {
	f.alias = alias
	return f
}

type FieldValue struct {
	value interface{}
	alias string
}

var _ Field = FieldValue{}

func (f FieldValue) GetAlias() string {
	if f.alias != "" {
		return f.alias
	}
	if v, ok := f.value.(interface{ GetAlias() string }); ok {
		return v.GetAlias()
	}
	return ""
}

func (f FieldValue) GetName() string { return "" }

func (f FieldValue) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	return BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, f.value, "")
}

func Value(value interface{}) FieldValue {
	return FieldValue{value: value}
}

func (f FieldValue) As(alias string) FieldValue {
	f.alias = alias
	return f
}

type Fields []Field

var _ SQLExcludeAppender = Fields{}

func (fs Fields) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	var err error
	for i, field := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		if field == nil {
			err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, nil, "")
			if err != nil {
				return fmt.Errorf("field #%d: %w", i+1, err)
			}
		} else {
			err = field.AppendSQLExclude(dialect, buf, args, params, env, excludedTableQualifiers)
			if err != nil {
				return fmt.Errorf("field #%d: %w", i+1, err)
			}
		}
	}
	return nil
}

type AliasFields []Field

func (fs AliasFields) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	var alias string
	var err error
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		if f == nil {
			err = BufferPrintValue(dialect, buf, args, params, env, excludedTableQualifiers, nil, "")
			if err != nil {
				return fmt.Errorf("field #%d: %w", i+1, err)
			}
		} else {
			err = f.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
			if err != nil {
				return fmt.Errorf("field #%d: %w", i+1, err)
			}
			if alias = f.GetAlias(); alias != "" {
				buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
			}
		}
	}
	return nil
}
