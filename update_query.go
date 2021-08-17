package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type UpdateQuery struct {
	Dialect      string
	Env          map[string]interface{}
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// UPDATE
	UpdateTable SchemaTable
	// FROM
	FromTable  Table
	JoinTables JoinTables
	// SET
	Assignments Assignments
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

var _ Query = UpdateQuery{}

func (q UpdateQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if env == nil && q.Env != nil {
		env = q.Env
	}
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
		err = q.CTEs.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("WITH: %w", err)
		}
	}
	// UPDATE
	buf.WriteString("UPDATE ")
	if q.UpdateTable == nil {
		return fmt.Errorf("no table provided to UPDATE")
	}
	err = q.UpdateTable.AppendSQL(dialect, buf, args, params, env)
	if err != nil {
		return fmt.Errorf("UPDATE: %w", err)
	}
	if alias := q.UpdateTable.GetAlias(); alias != "" {
		buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
		excludedTableQualifiers = append(excludedTableQualifiers, alias)
	} else {
		name := q.UpdateTable.GetName()
		excludedTableQualifiers = append(excludedTableQualifiers, name)
	}
	if len(q.Assignments) == 0 {
		return fmt.Errorf("no fields to update")
	}
	// SET (not MySQL)
	if dialect != DialectMySQL {
		buf.WriteString(" SET ")
		err = q.Assignments.AppendSQLExclude(dialect, buf, args, params, env, excludedTableQualifiers)
		if err != nil {
			return fmt.Errorf("SET: %w", err)
		}
	}
	// FROM
	if q.FromTable != nil {
		if dialect == DialectMySQL {
			return fmt.Errorf("mysql UPDATE does not support FROM")
		}
		buf.WriteString(" FROM ")
		err = q.FromTable.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("FROM: %w", err)
		}
		alias := q.FromTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
		}
	}
	// JOIN
	if len(q.JoinTables) > 0 {
		if q.FromTable == nil && dialect != DialectMySQL {
			return fmt.Errorf("%s can't JOIN without a FROM table", dialect)
		}
		buf.WriteString(" ")
		err = q.JoinTables.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("JOIN: %w", err)
		}
	}
	// SET (MySQL)
	if len(q.Assignments) > 0 && dialect == DialectMySQL {
		buf.WriteString(" SET ")
		err = q.Assignments.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("SET: %w", err)
		}
	}
	// WHERE
	var wherePredicate VariadicPredicate
	if predicateInjector, ok := q.UpdateTable.(PredicateHook); ok {
		predicate, err := predicateInjector.InjectPredicate(env)
		if err != nil {
			return fmt.Errorf("table %s injecting predicate: %w", q.UpdateTable.GetName(), err)
		}
		if predicate != nil {
			wherePredicate.Predicates = append(wherePredicate.Predicates, predicate)
		}
	}
	if predicateInjector, ok := q.FromTable.(PredicateHook); ok {
		predicate, err := predicateInjector.InjectPredicate(env)
		if err != nil {
			return fmt.Errorf("table %s injecting predicate: %w", q.FromTable.GetName(), err)
		}
		if predicate != nil {
			wherePredicate.Predicates = append(wherePredicate.Predicates, predicate)
		}
	}
	for i, joinTable := range q.JoinTables {
		if predicateInjector, ok := joinTable.Table.(PredicateHook); ok {
			predicate, err := predicateInjector.InjectPredicate(env)
			if err != nil {
				return fmt.Errorf("table #%d %s injecting predicate: %w", i+1, joinTable.Table.GetName(), err)
			}
			if predicate != nil {
				wherePredicate.Predicates = append(wherePredicate.Predicates, predicate)
			}
		}
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
			return fmt.Errorf("%s UPDATE does not support ORDER BY", dialect)
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
			return fmt.Errorf("%s UPDATE does not support LIMIT", dialect)
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
			return fmt.Errorf("%s UPDATE does not support RETURNING", dialect)
		}
		buf.WriteString(" RETURNING ")
		err = q.ReturningFields.AppendSQLExclude(dialect, buf, args, params, env, nil)
		if err != nil {
			return fmt.Errorf("RETURNING: %w", err)
		}
	}
	return nil
}

func (q UpdateQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.Dialect {
	case DialectPostgres, DialectSQLite:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, fmt.Errorf("%s UPDATE %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q UpdateQuery) GetFetchableFields() ([]Field, error) {
	switch q.Dialect {
	case DialectPostgres, DialectSQLite:
		return q.ReturningFields, nil
	default:
		return nil, fmt.Errorf("%s UPDATE %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q UpdateQuery) GetDialect() string { return q.Dialect }
