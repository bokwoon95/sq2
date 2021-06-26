package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type DeleteQuery struct {
	QueryDialect string
	// WITH
	CTEs CTEs
	// DELETE FROM
	FromTable BaseTable
	// USING
	UsingTable Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	QueryLimit sql.NullInt64
	// RETURNING
	ReturningFields Fields
}

var _ Query = DeleteQuery{}

func (q DeleteQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var err error
	// WITH
	if len(q.CTEs) > 0 {
		err = q.CTEs.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	// DELETE FROM
	buf.WriteString("DELETE FROM ")
	if q.FromTable == nil {
		return fmt.Errorf("DELETE-ing from a nil table")
	}
	err = q.FromTable.AppendSQL(dialect, buf, args, params)
	if err != nil {
		return err
	}
	if alias := q.FromTable.GetAlias(); alias != "" {
		buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
	}
	// USING
	if q.UsingTable != nil && (dialect == DialectPostgres || dialect == DialectMySQL) {
		buf.WriteString(" USING ")
		err = q.UsingTable.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
		if alias := q.UsingTable.GetAlias(); alias != "" {
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
	if len(q.OrderByFields) > 0 {
		buf.WriteString(" ORDER BY ")
		err = q.OrderByFields.AppendSQLExclude(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	// LIMIT
	if q.QueryLimit.Valid {
		err = BufferPrintf(dialect, buf, args, params, nil, " LIMIT {}", []interface{}{q.QueryLimit.Int64})
		if err != nil {
			return err
		}
	}
	// RETURNING
	if len(q.ReturningFields) > 0 {
		buf.WriteString(" RETURNING ")
		err = q.ReturningFields.AppendSQLExcludeWithAlias(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q DeleteQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.QueryDialect {
	case DialectPostgres:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, fmt.Errorf("%s DELETE %w", q.QueryDialect, ErrNonFetchableQuery)
	}
}

func (q DeleteQuery) GetFetchableFields() ([]Field, error) {
	switch q.QueryDialect {
	case DialectPostgres:
		return q.ReturningFields, nil
	default:
		return nil, fmt.Errorf("%s DELETE %w", q.QueryDialect, ErrNonFetchableQuery)
	}
}

func (q DeleteQuery) Dialect() string { return q.QueryDialect }
