package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type SelectQuery struct {
	Dialect string
	Env     map[string]interface{}
	// WITH
	CTEs CTEs
	// SELECT
	Distinct         bool
	SelectFields     AliasFields
	DistinctOnFields Fields
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
	RowLimit sql.NullInt64
	// OFFSET
	RowOffset sql.NullInt64
}

var _ Query = SelectQuery{}

func (q SelectQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if env == nil && q.Env != nil {
		env = q.Env
	}
	var err error
	// WITH
	if len(q.CTEs) > 0 {
		err = q.CTEs.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("WITH: %w", err)
		}
	}
	// SELECT
	buf.WriteString("SELECT ")
	if len(q.DistinctOnFields) > 0 {
		if dialect != DialectPostgres {
			return fmt.Errorf("%s does not support SELECT DISTINCT ON", dialect)
		}
		if q.Distinct {
			return fmt.Errorf("postgres SELECT cannot be DISTINCT and DISTINCT ON at the same time")
		}
		buf.WriteString("DISTINCT ON (")
		err = q.DistinctOnFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("DISTINCT ON: %w", err)
		}
		buf.WriteString(") ")
	} else if q.Distinct {
		buf.WriteString("DISTINCT ")
	}
	if len(q.SelectFields) == 0 {
		return fmt.Errorf("no fields SELECT-ed")
	}
	err = q.SelectFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
	if err != nil {
		return fmt.Errorf("SELECT: %w", err)
	}
	// FROM
	if q.FromTable != nil {
		buf.WriteString(" FROM ")
		err = q.FromTable.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("FROM: %w", err)
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
		err = q.JoinTables.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("JOIN: %w", err)
		}
	}
	// WHERE
	var wherePredicate VariadicPredicate
	if predicateHook, ok := q.FromTable.(PredicateHook); ok {
		predicate, err := predicateHook.InjectPredicate(env)
		if err != nil {
			return fmt.Errorf("table %s injecting predicate: %w", q.FromTable.GetName(), err)
		}
		if predicate != nil {
			wherePredicate.Predicates = append(wherePredicate.Predicates, predicate)
		}
	}
	for i, joinTable := range q.JoinTables {
		if predicateHook, ok := joinTable.Table.(PredicateHook); ok {
			predicate, err := predicateHook.InjectPredicate(env)
			if err != nil {
				return fmt.Errorf("table #%d %s injecting predicate: %w", i+1, joinTable.Table.GetName(), err)
			}
			if predicate != nil {
				wherePredicate.Predicates = append(wherePredicate.Predicates, predicate)
			}
		}
	}
	if len(wherePredicate.Predicates) > 0 {
		wherePredicate.Predicates = append(wherePredicate.Predicates, q.WherePredicate)
		q.WherePredicate = wherePredicate
	}
	if len(q.WherePredicate.Predicates) > 0 {
		buf.WriteString(" WHERE ")
		q.WherePredicate.Toplevel = true
		err = q.WherePredicate.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("WHERE: %w", err)
		}
	}
	// GROUP BY
	if len(q.GroupByFields) > 0 {
		buf.WriteString(" GROUP BY ")
		err = q.GroupByFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("GROUP BY: %w", err)
		}
	}
	// HAVING
	if len(q.HavingPredicate.Predicates) > 0 {
		buf.WriteString(" HAVING ")
		q.HavingPredicate.Toplevel = true
		err = q.HavingPredicate.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("HAVING: %w", err)
		}
	}
	// ORDER BY
	if len(q.OrderByFields) > 0 {
		buf.WriteString(" ORDER BY ")
		err = q.OrderByFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("ORDER BY: %w", err)
		}
	}
	// LIMIT
	if q.RowLimit.Valid {
		err = BufferPrintf(dialect, buf, args, params, env, nil, " LIMIT {}", []interface{}{q.RowLimit.Int64})
		if err != nil {
			return fmt.Errorf("LIMIT: %w", err)
		}
	}
	// OFFSET
	if q.RowOffset.Valid {
		err = BufferPrintf(dialect, buf, args, params, env, nil, " OFFSET {}", []interface{}{q.RowOffset.Int64})
		if err != nil {
			return fmt.Errorf("OFFSET: %w", err)
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

func (q SelectQuery) GetDialect() string { return q.Dialect }
