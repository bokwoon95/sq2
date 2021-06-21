package sq

import (
	"bytes"
	"fmt"
)

type SubqueryField struct {
	fetchFieldsErr error
	query          Query
	QueryAlias     string
	FieldName      string
	FieldAlias     string
}

var _ Field = SubqueryField{}

func (f SubqueryField) GetAlias() string { return f.FieldAlias }

func (f SubqueryField) GetName() string { return f.FieldName }

func (f SubqueryField) As(alias string) SubqueryField {
	f.FieldAlias = alias
	return f
}

func (f SubqueryField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if f.fetchFieldsErr != nil {
		return fmt.Errorf("sq: unable to retrieve subquery's fields: %w", f.fetchFieldsErr)
	}
	if f.QueryAlias == "" && f.FieldName == "" {
		return fmt.Errorf("sq: referenced nonexistent SubqueryField")
	}
	if f.FieldName == "" {
		return fmt.Errorf("sq: no named assigned to SubqueryField")
	}
	buf.WriteString(f.QueryAlias + "." + f.FieldName)
	return nil
}

type Subquery map[string]SubqueryField

var _ Table = Subquery{}

func NewSubquery(query Query, alias string) Subquery {
	q := map[string]SubqueryField{}
	metafield := SubqueryField{query: query, QueryAlias: alias}
	var fields []Field
	fields, metafield.fetchFieldsErr = query.GetFetchableFields()
	if metafield.fetchFieldsErr != nil {
		return q
	}
	for _, field := range fields {
		var subqueryField SubqueryField
		subqueryField.QueryAlias = alias
		subqueryField.FieldName = field.GetAlias()
		if subqueryField.FieldName == "" {
			subqueryField.FieldName = field.GetName()
		}
		q[subqueryField.FieldName] = subqueryField
	}
	return q
}

func (q Subquery) GetName() string { return "" }

func (q Subquery) GetAlias() string { return q[""].QueryAlias }

func (q Subquery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	query := q[""].query
	if query == nil {
		return fmt.Errorf("empty subquery")
	}
	err := query.AppendSQL(dialect, buf, args, params)
	if err != nil {
		return err
	}
	return nil
}
