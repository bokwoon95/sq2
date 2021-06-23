package sq

import (
	"bytes"
	"fmt"
)

type InsertQuery struct {
	QueryDialect string
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// NOTE: added MySQL's INSERT IGNORE functionality but disabled it for now
	// because I'm not sure if INSERT IGNORE is something I want to encourage.
	// https://stackoverflow.com/questions/548541/insert-ignore-vs-insert-on-duplicate-key-update
	ignore bool
	// INSERT INTO
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
	if q.ignore && dialect == DialectMySQL {
		buf.WriteString("INSERT IGNORE INTO ")
	} else {
		buf.WriteString("INSERT INTO ")
	}
	if q.IntoTable == nil {
		return fmt.Errorf("sq: INSERTing into nil table")
	}
	err := q.IntoTable.AppendSQL(dialect, buf, args, params)
	if err != nil {
		return err
	}
	if alias := q.IntoTable.GetAlias(); alias != "" {
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
		return fmt.Errorf("RowValues not provided and SelectQuery not provided")
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
				return fmt.Errorf("no conflict target specified")
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
	if len(q.ReturningFields) > 0 && dialect == DialectPostgres {
		buf.WriteString(" RETURNING ")
		err = q.ReturningFields.AppendSQLExcludeWithAlias(dialect, buf, args, params, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q InsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.QueryDialect {
	case DialectPostgres:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, ErrUnsupported
	}
}

func (q InsertQuery) GetFetchableFields() ([]Field, error) {
	switch q.QueryDialect {
	case DialectPostgres:
		return q.ReturningFields, nil
	default:
		return nil, ErrUnsupported
	}
}

func (q InsertQuery) Dialect() string { return q.QueryDialect }
