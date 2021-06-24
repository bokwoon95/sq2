package sq

import (
	"bytes"
	"fmt"
)

type SubqueryField struct {
	valid      bool
	query      Query
	queryAlias string
	fieldName  string
	fieldAlias string
}

var _ Field = SubqueryField{}

func (f SubqueryField) GetAlias() string { return f.fieldAlias }

func (f SubqueryField) GetName() string { return f.fieldName }

func (f SubqueryField) As(alias string) SubqueryField {
	f.fieldAlias = alias
	return f
}

func (f SubqueryField) AppendSQLExclude(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, excludedTableQualifiers []string) error {
	if !f.valid {
		return fmt.Errorf("referenced nonexistent SubqueryField")
	}
	if f.fieldName == "" {
		return fmt.Errorf("SubqueryField has no name")
	}
	if f.queryAlias != "" {
		buf.WriteString(f.queryAlias + ".")
	}
	buf.WriteString(f.fieldName)
	return nil
}

type Subquery map[string]SubqueryField

var _ Table = Subquery{}

func NewSubquery(query Query, alias string) (Subquery, error) {
	if query == nil {
		return nil, fmt.Errorf("Subquery query cannot be nil")
	}
	q := Subquery{"": {
		query:      query,
		queryAlias: alias,
	}}
	fields, err := query.GetFetchableFields()
	if err != nil {
		return nil, fmt.Errorf("error fetching fields for Subquery: %w", err)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("Subquery query does not return any fields")
	}
	for i, field := range fields {
		if field == nil {
			return nil, fmt.Errorf("field #%d in Subquery query is nil", i+1)
		}
		fieldName := field.GetAlias()
		if fieldName == "" {
			fieldName = field.GetName()
		}
		if fieldName == "" {
			return q, fmt.Errorf("field #%d in Subquery has no name and no alias", i+1)
		}
		q[fieldName] = SubqueryField{
			valid:      true,
			queryAlias: alias,
			fieldName:  fieldName,
		}
	}
	return q, nil
}

func (q Subquery) GetAlias() string { return q[""].queryAlias }

func (q Subquery) GetName() string { return "" }

func (q Subquery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if len(q) == 0 {
		return fmt.Errorf("empty Subquery")
	}
	query := q[""].query
	if query == nil {
		return fmt.Errorf("empty Subquery")
	}
	buf.WriteString("(")
	err := query.AppendSQL(dialect, buf, args, params)
	if err != nil {
		return err
	}
	buf.WriteString(")")
	return nil
}
