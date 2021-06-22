package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type UpdateQuery struct {
	QueryDialect string
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// UPDATE
	UpdateTables []BaseTable
	// SET
	Assignments Assignments
	// FROM
	FromTable  Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
	// RETURNING
	ReturningFields Fields
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	QueryLimit sql.NullInt64
}

var _ Query = UpdateQuery{}

func (q UpdateQuery) ToSQL() (query string, args []interface{}, params map[string][]int, err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string][]int)
	err = q.AppendSQL(q.QueryDialect, buf, &args, params)
	if err != nil {
		return query, args, params, err
	}
	query = buf.String()
	return query, args, params, nil
}

func (q UpdateQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var err error
	var excludedTableQualifiers []string
	if q.ColumnMapper != nil {
		col := NewColumn(ColumnModeUpdate)
		err := q.ColumnMapper(col)
		if err != nil {
			return err
		}
		q.Assignments = ColumnUpdateResult(col)
	}
	// WITH
	if len(q.CTEs) > 0 {
		err = q.CTEs.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	// UPDATE
	buf.WriteString("UPDATE ")
	if len(q.UpdateTables) == 0 {
		return fmt.Errorf("sq: no tables to UPDATE")
	}
	// TODO: holy shit I'm not smart enough for MySQL's multi-table update semantics
	for i, updateTable := range q.UpdateTables {
		if i > 0 {
			if dialect != DialectMySQL {
				break
			}
			buf.WriteString(", ")
		}
		if updateTable == nil {
			return fmt.Errorf("sq: UPDATE-ing a nil table")
		}
		err = updateTable.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
		if alias := updateTable.GetAlias(); alias != "" {
			buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
			excludedTableQualifiers = append(excludedTableQualifiers, alias)
		} else {
			name := updateTable.GetName()
			excludedTableQualifiers = append(excludedTableQualifiers, name)
		}
	}
	// SET
	if len(q.Assignments) > 0 {
		buf.WriteString(" SET ")
		err = q.Assignments.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
		if err != nil {
			return err
		}
	}
	// FROM
	if q.FromTable != nil && dialect != DialectMySQL {
		buf.WriteString(" FROM ")
		err = q.FromTable.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
		alias := q.FromTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
		}
	}
	// JOIN
	if len(q.JoinTables) > 0 {
		buf.WriteString(" ")
		err = q.JoinTables.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	// WHERE
	if len(q.WherePredicate.Predicates) > 0 {
		buf.WriteString(" WHERE ")
		q.WherePredicate.Toplevel = true
		err = q.WherePredicate.AppendSQLExclude(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	// ORDER BY
	if len(q.OrderByFields) > 0 && dialect == DialectMySQL {
		buf.WriteString(" ORDER BY ")
		err = q.OrderByFields.AppendSQLExclude(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	// LIMIT
	if q.QueryLimit.Valid && dialect == DialectMySQL {
		err = BufferPrintf(dialect, buf, args, params, nil, " LIMIT {}", []interface{}{q.QueryLimit.Int64})
		if err != nil {
			return err
		}
	}
	// RETURNING
	if len(q.ReturningFields) > 0 && dialect == DialectPostgres {
		buf.WriteString(" RETURNING ")
		q.ReturningFields.AppendSQLExcludeWithAlias(dialect, buf, args, params, nil)
	}
	return nil
}

func (q UpdateQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.QueryDialect {
	case DialectPostgres:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, ErrUnsupported
	}
}

func (q UpdateQuery) GetFetchableFields() ([]Field, error) {
	switch q.QueryDialect {
	case DialectPostgres:
		return q.ReturningFields, nil
	default:
		return nil, ErrUnsupported
	}
}

func (q UpdateQuery) Dialect() string { return q.QueryDialect }
