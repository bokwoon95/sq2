package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type SelectQuery struct {
	QueryDialect string
	// WITH
	CTEs CTEs
	// SELECT
	SelectType   SelectType
	SelectFields Fields
	DistinctOn   Fields
	// FROM
	FromTable  Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
	// GROUP BY
	GroupByFields Fields
	// HAVING
	HavingPredicate VariadicPredicate
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	QueryLimit sql.NullInt64
	// OFFSET
	QueryOffset sql.NullInt64
}

var _ Query = SelectQuery{}

func (q SelectQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var err error
	// WITH
	if len(q.CTEs) > 0 {
		err = q.CTEs.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	// SELECT
	if q.SelectType == "" {
		q.SelectType = SelectTypeDefault
	}
	buf.WriteString(string(q.SelectType))
	if len(q.SelectFields) == 0 {
		return fmt.Errorf("no fields SELECT-ed")
	}
	buf.WriteString(" ")
	err = q.SelectFields.AppendSQLExcludeWithAlias(dialect, buf, args, params, nil)
	if err != nil {
		return err
	}
	// FROM
	if q.FromTable != nil {
		buf.WriteString(" FROM ")
		err = q.FromTable.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
		if tableAlias := q.FromTable.GetAlias(); tableAlias != "" {
			buf.WriteString(" AS " + QuoteIdentifier(dialect, tableAlias))
		}
	}
	// JOIN
	if len(q.JoinTables) > 0 {
		if q.FromTable == nil {
			return fmt.Errorf("can't JOIN without a FROM table")
		}
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
	// GROUP BY
	if len(q.GroupByFields) > 0 {
		buf.WriteString(" GROUP BY ")
		err = q.GroupByFields.AppendSQLExclude(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	// HAVING
	if len(q.HavingPredicate.Predicates) > 0 {
		buf.WriteString(" HAVING ")
		q.HavingPredicate.Toplevel = true
		err = q.HavingPredicate.AppendSQLExclude(dialect, buf, args, params, nil)
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
	// OFFSET
	if q.QueryOffset.Valid {
		err = BufferPrintf(dialect, buf, args, params, nil, " OFFSET {}", []interface{}{q.QueryOffset.Int64})
		if err != nil {
			return err
		}
	}
	return nil
}

func (q SelectQuery) SetFetchableFields(fields []Field) (Query, error) {
	q.SelectFields = fields
	return q, nil
}

func (q SelectQuery) GetFetchableFields() ([]Field, error) {
	return q.SelectFields, nil
}

func (q SelectQuery) Dialect() string { return q.QueryDialect }
