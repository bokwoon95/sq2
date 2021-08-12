package sq

import (
	"bytes"
	"fmt"
	"strings"
)

type Subquery struct {
	query          Query
	stickyErr      error
	explicitFields bool
	subqueryAlias  string
	fieldNames     []string
	fieldCache     map[string]int
}

var _ Table = Subquery{}

func NewSubquery(alias string, query Query) Subquery {
	q := Subquery{
		query:         query,
		subqueryAlias: alias,
		fieldCache:    make(map[string]int),
	}
	if query == nil {
		q.stickyErr = fmt.Errorf("subquery query cannot be nil")
		return q
	}
	fields, err := query.GetFetchableFields()
	if err != nil {
		q.stickyErr = fmt.Errorf("subquery %s failed to fetch query fields: %w", q.subqueryAlias, err)
		return q
	}
	if len(fields) == 0 {
		q.stickyErr = fmt.Errorf("subquery %s does not return any fields", q.subqueryAlias)
		return q
	}
	for i, field := range fields {
		fieldName := field.GetAlias()
		if fieldName == "" {
			fieldName = field.GetName()
		}
		if fieldName == "" {
			q.stickyErr = fmt.Errorf("subquery %s field #%d needs a name or an alias", q.subqueryAlias, i+1)
			return q
		}
		q.fieldNames = append(q.fieldNames, fieldName)
	}
	for i, fieldName := range q.fieldNames {
		q.fieldCache[fieldName] = i
	}
	return q
}

func (q Subquery) GetAlias() string { return q.subqueryAlias }

func (q Subquery) GetName() string { return "" }

func (q Subquery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if q.stickyErr != nil {
		return q.stickyErr
	}
	buf.WriteString("(")
	err := q.query.AppendSQL(dialect, buf, args, params, nil)
	if err != nil {
		return fmt.Errorf("subquery failed to build query: %w", err)
	}
	buf.WriteString(")")
	if q.subqueryAlias == "" && (dialect == DialectPostgres || dialect == DialectMySQL) {
		return fmt.Errorf("%s subquery needs an alias", dialect)
	}
	return nil
}

func (q Subquery) Field(fieldName string) SubqueryField {
	return SubqueryField{
		stickyErr:  q.stickyErr,
		fieldNames: q.fieldNames,
		fieldCache: q.fieldCache,
		info: FieldInfo{
			TableAlias: q.subqueryAlias,
			FieldName:  fieldName,
		},
	}
}

type SubqueryField struct {
	stickyErr  error
	fieldNames []string
	fieldCache map[string]int
	info       FieldInfo
}

var _ Field = SubqueryField{}

func (f SubqueryField) GetAlias() string { return f.info.FieldAlias }

func (f SubqueryField) GetName() string { return f.info.FieldName }

func (f SubqueryField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}, excludedTableQualifiers []string) error {
	if f.info.TableAlias == "" && (dialect == DialectPostgres || dialect == DialectMySQL) {
		return fmt.Errorf("subquery field %s invalid because %s subquery needs an alias", f.info.FieldName, dialect)
	}
	if _, ok := f.fieldCache[f.info.FieldName]; !ok {
		if f.stickyErr != nil {
			return fmt.Errorf("subquery field %s.%s invalid due to Subquery error: %w", f.info.TableAlias, f.info.FieldName, f.stickyErr)
		} else {
			return fmt.Errorf("subquery field %s.%s does not exist (available fields: %s)", f.info.TableAlias, f.info.FieldName, strings.Join(f.fieldNames, ", "))
		}
	}
	return f.info.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
}

func (f SubqueryField) As(alias string) SubqueryField {
	f.info.FieldAlias = alias
	return f
}

func (f SubqueryField) Asc() SubqueryField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = false
	return f
}

func (f SubqueryField) Desc() SubqueryField {
	f.info.Descending.Valid = true
	f.info.Descending.Bool = true
	return f
}

func (f SubqueryField) NullsLast() SubqueryField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = false
	return f
}

func (f SubqueryField) NullsFirst() SubqueryField {
	f.info.NullsFirst.Valid = true
	f.info.NullsFirst.Bool = true
	return f
}

func (f SubqueryField) IsNull() Predicate { return IsNull(f) }

func (f SubqueryField) IsNotNull() Predicate { return IsNotNull(f) }

func (f SubqueryField) In(v interface{}) Predicate { return In(f, v) }

func (f SubqueryField) Eq(v interface{}) Predicate { return Eq(f, v) }

func (f SubqueryField) Ne(v interface{}) Predicate { return Ne(f, v) }

func (f SubqueryField) Gt(v interface{}) Predicate { return Gt(f, v) }

func (f SubqueryField) Ge(v interface{}) Predicate { return Ge(f, v) }

func (f SubqueryField) Lt(v interface{}) Predicate { return Lt(f, v) }

func (f SubqueryField) Le(v interface{}) Predicate { return Le(f, v) }
