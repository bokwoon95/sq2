package sq

import (
	"bytes"
	"database/sql"
	"fmt"
)

type DeleteQuery struct {
	Dialect string
	// WITH
	CTEs CTEs
	// DELETE FROM
	FromTables []BaseTable
	// USING
	UsingTable Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	RowLimit sql.NullInt64
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
		err = fromTable.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
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
		if dialect == DialectSQLite {
			return fmt.Errorf("sqlite DELETE does not support joins")
		}
		if q.UsingTable == nil {
			return fmt.Errorf("can't use JOIN without providing an initial table to join on")
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
	// ORDER BY
	if len(q.OrderByFields) > 0 {
		if dialect != DialectMySQL {
			return fmt.Errorf("%s DELETE does not support ORDER BY", dialect)
		}
		if q.UsingTable != nil {
			return fmt.Errorf("ORDER BY not allowed in a multi-table DELETE")
		}
		buf.WriteString(" ORDER BY ")
		err = q.OrderByFields.AppendSQLExclude(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	// LIMIT
	if q.RowLimit.Valid {
		if dialect != DialectMySQL {
			return fmt.Errorf("%s DELETE does not support LIMIT", dialect)
		}
		if q.UsingTable != nil {
			return fmt.Errorf("LIMIT not allowed in a multi-table DELETE")
		}
		err = BufferPrintf(dialect, buf, args, params, nil, " LIMIT {}", []interface{}{q.RowLimit.Int64})
		if err != nil {
			return err
		}
	}
	// RETURNING
	if len(q.ReturningFields) > 0 {
		if dialect != DialectPostgres && dialect != DialectSQLite {
			return fmt.Errorf("%s DELETE does not support RETURNING", dialect)
		}
		buf.WriteString(" RETURNING ")
		err = q.ReturningFields.AppendSQLExcludeWithAlias(dialect, buf, args, params, nil)
		if err != nil {
			return err
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
