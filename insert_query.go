package sq

import (
	"bytes"
	"fmt"
)

type InsertQuery struct {
	Dialect      string
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// INSERT INTO
	InsertIgnore  bool
	IntoTable     BaseTable
	InsertColumns Fields
	// VALUES
	RowValues RowValues
	// SELECT
	SelectQuery *SelectQuery
	// ON CONFLICT
	HandleConflict      bool
	ConflictFields      Fields
	ConflictPredicate   VariadicPredicate
	ConflictConstraint  string
	Resolution          Assignments
	ResolutionPredicate VariadicPredicate
	// RETURNING
	ReturningFields Fields
}

var _ Query = InsertQuery{}

func (q InsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var excludedTableQualifiers []string
	if q.ColumnMapper != nil {
		col := NewColumn(ColumnModeInsert)
		err := q.ColumnMapper(col)
		if err != nil {
			return err
		}
		q.InsertColumns, q.RowValues = ColumnInsertResult(col)
	}
	// WITH
	if len(q.CTEs) > 0 {
		err := q.CTEs.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	// INSERT INTO
	if q.InsertIgnore {
		if dialect != DialectMySQL {
			return fmt.Errorf("%s does not support INSERT IGNORE", dialect)
		}
		buf.WriteString("INSERT IGNORE INTO ")
	} else {
		buf.WriteString("INSERT INTO ")
	}
	if q.IntoTable == nil {
		return fmt.Errorf("no table provided to INSERT")
	}
	err := q.IntoTable.AppendSQL(dialect, buf, args, params)
	if err != nil {
		return err
	}
	if alias := q.IntoTable.GetAlias(); alias != "" {
		if dialect == DialectMySQL {
			return fmt.Errorf("mysql does not allow an alias for the INSERT table")
		}
		buf.WriteString(" AS " + QuoteIdentifier(dialect, alias))
		excludedTableQualifiers = append(excludedTableQualifiers, alias)
	} else {
		name := q.IntoTable.GetName()
		excludedTableQualifiers = append(excludedTableQualifiers, name)
	}
	if len(q.InsertColumns) > 0 {
		buf.WriteString(" (")
		err = q.InsertColumns.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
		if err != nil {
			return err
		}
		buf.WriteString(")")
	}
	// VALUES/SELECT
	switch {
	case len(q.RowValues) > 0:
		buf.WriteString(" VALUES ")
		err = q.RowValues.AppendSQL(dialect, buf, args, nil)
		if err != nil {
			return err
		}
	case q.SelectQuery != nil:
		buf.WriteString(" ")
		err = q.SelectQuery.AppendSQL(dialect, buf, args, nil)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("RowValues not provided and SelectQuery not provided to INSERT query")
	}
	// ON CONFLICT
	switch dialect {
	case DialectSQLite, DialectPostgres:
		if q.HandleConflict {
			buf.WriteString(" ON CONFLICT")
			if q.ConflictConstraint != "" && dialect == DialectPostgres {
				buf.WriteString(" ON CONSTRAINT " + q.ConflictConstraint)
			} else if len(q.ConflictFields) > 0 {
				buf.WriteString(" (")
				err = q.ConflictFields.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
				if err != nil {
					return err
				}
				buf.WriteString(")")
				if len(q.ConflictPredicate.Predicates) > 0 {
					buf.WriteString(" WHERE ")
					q.ConflictPredicate.Toplevel = true
					err = q.ConflictPredicate.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
					if err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("INSERT query has no conflict target specified")
			}
		}
		if q.HandleConflict && len(q.Resolution) > 0 {
			buf.WriteString(" DO UPDATE SET ")
			err = q.Resolution.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
			if err != nil {
				return err
			}
			if len(q.ResolutionPredicate.Predicates) > 0 {
				buf.WriteString(" WHERE ")
				q.ResolutionPredicate.Toplevel = true
				err = q.ResolutionPredicate.AppendSQLExclude(dialect, buf, args, params, nil)
				if err != nil {
					return err
				}
			}
		}
		if q.HandleConflict && len(q.Resolution) == 0 {
			buf.WriteString(" DO NOTHING")
		}
	case DialectMySQL:
		if len(q.Resolution) > 0 {
			buf.WriteString(" ON DUPLICATE KEY UPDATE ")
			err = q.Resolution.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
			if err != nil {
				return err
			}
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

func (q InsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.Dialect {
	case DialectPostgres:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, fmt.Errorf("%s INSERT %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q InsertQuery) GetFetchableFields() ([]Field, error) {
	switch q.Dialect {
	case DialectPostgres:
		return q.ReturningFields, nil
	default:
		return nil, fmt.Errorf("%s INSERT %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q InsertQuery) GetDialect() string { return q.Dialect }
