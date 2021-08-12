package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type DeleteQuery struct {
	Dialect string
	Env     map[string]interface{}
	// WITH
	CTEs CTEs
	// DELETE FROM
	FromTables []SchemaTable
	// USING
	UsingTable Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	RowLimit sql.NullInt64
	// OFFSET
	RowOffset sql.NullInt64
	// RETURNING
	ReturningFields AliasFields
}

var _ Query = DeleteQuery{}

func (q DeleteQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
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
	// DELETE FROM
	buf.WriteString("DELETE FROM ")
	if len(q.FromTables) == 0 {
		return fmt.Errorf("no table provided to DELETE")
	}
	if q.UsingTable != nil && dialect == DialectMySQL {
		for i, table := range q.FromTables {
			if i > 0 {
				buf.WriteString(", ")
			}
			nameOrAlias := table.GetAlias()
			if nameOrAlias == "" {
				nameOrAlias = table.GetName()
			}
			buf.WriteString(nameOrAlias)
		}
	} else {
		fromTable := q.FromTables[0]
		if fromTable == nil {
			return fmt.Errorf("no table provided to DELETE")
		}
		err = fromTable.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("DELETE FROM: %w", err)
		}
		if alias := fromTable.GetAlias(); alias != "" {
			buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
		}
	}
	// USING
	if q.UsingTable != nil {
		if dialect == DialectSQLite {
			return fmt.Errorf("sqlite DELETE does not support joins")
		}
		buf.WriteString(" USING ")
		err = q.UsingTable.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("USING: %w", err)
		}
		if alias := q.UsingTable.GetAlias(); alias != "" {
			buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
		}
	}
	// JOIN
	if len(q.JoinTables) > 0 {
		if dialect == DialectSQLite {
			return fmt.Errorf("sqlite DELETE does not support joins")
		}
		if q.UsingTable == nil {
			return fmt.Errorf("can't use JOIN without providing an initial table to join on")
		}
		buf.WriteString(" ")
		err = q.JoinTables.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("JOIN: %w", err)
		}
	}
	// WHERE
	var tablePredicates []Predicate
	for i, table := range q.FromTables {
		if predicateAdder, ok := table.(PredicateAdder); ok {
			predicates, err := predicateAdder.AddPredicate(env)
			if err != nil {
				return fmt.Errorf("table #%d adding predicate: %w", i+1, err)
			}
			tablePredicates = append(tablePredicates, predicates...)
		}
	}
	if len(tablePredicates) > 0 {
		q.WherePredicate.Predicates = append(tablePredicates, q.WherePredicate.Predicates...)
	}
	if len(q.WherePredicate.Predicates) > 0 {
		buf.WriteString(" WHERE ")
		q.WherePredicate.Toplevel = true
		err = q.WherePredicate.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("WHERE: %w", err)
		}
	}
	// ORDER BY
	if len(q.OrderByFields) > 0 {
		if dialect != DialectMySQL && dialect != DialectSQLite {
			return fmt.Errorf("%s DELETE does not support ORDER BY", dialect)
		}
		if q.UsingTable != nil {
			return fmt.Errorf("ORDER BY not allowed in a multi-table DELETE")
		}
		buf.WriteString(" ORDER BY ")
		err = q.OrderByFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("ORDER BY: %w", err)
		}
	}
	// LIMIT
	if q.RowLimit.Valid {
		if dialect != DialectMySQL && dialect != DialectSQLite {
			return fmt.Errorf("%s DELETE does not support LIMIT", dialect)
		}
		if q.UsingTable != nil {
			return fmt.Errorf("LIMIT not allowed in a multi-table DELETE")
		}
		err = BufferPrintf(dialect, buf, args, params, env, nil, " LIMIT {}", []interface{}{q.RowLimit.Int64})
		if err != nil {
			return fmt.Errorf("LIMIT: %w", err)
		}
	}
	// OFFSET
	if q.RowOffset.Valid {
		if dialect != DialectSQLite {
			return fmt.Errorf("%s UPDATE does not support OFFSET", dialect)
		}
		err = BufferPrintf(dialect, buf, args, params, env, nil, " OFFSET {}", []interface{}{q.RowOffset.Int64})
		if err != nil {
			return fmt.Errorf("OFFSET: %w", err)
		}
	}
	// RETURNING
	if len(q.ReturningFields) > 0 {
		if dialect != DialectPostgres && dialect != DialectSQLite {
			return fmt.Errorf("%s DELETE does not support RETURNING", dialect)
		}
		buf.WriteString(" RETURNING ")
		err = q.ReturningFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("RETURNING: %w", err)
		}
	}
	return nil
}

func (q DeleteQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.Dialect {
	case DialectPostgres:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, fmt.Errorf("%s DELETE %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q DeleteQuery) GetFetchableFields() ([]Field, error) {
	switch q.Dialect {
	case DialectPostgres:
		return q.ReturningFields, nil
	default:
		return nil, fmt.Errorf("%s DELETE %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q DeleteQuery) GetDialect() string { return q.Dialect }
