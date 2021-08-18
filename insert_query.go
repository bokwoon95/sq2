package sq

import (
	"bytes"
	"fmt"
	"strings"
)

type InsertQuery struct {
	Dialect      string
	Env          map[string]interface{}
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// INSERT INTO
	InsertIgnore  bool
	IntoTable     SchemaTable
	InsertColumns Fields
	// VALUES
	RowValues     RowValues
	RowAlias      string
	ColumnAliases []string
	// SELECT
	SelectQuery *SelectQuery
	// ON CONFLICT
	ConflictConstraint  string
	ConflictFields      Fields
	ConflictPredicate   Predicate
	ConflictDoNothing   bool
	Resolution          Assignments
	ResolutionPredicate VariadicPredicate
	// RETURNING
	ReturningFields AliasFields
}

var _ Query = (*InsertQuery)(nil)

func (q InsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	if env == nil && q.Env != nil {
		env = q.Env
	}
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
		err := q.CTEs.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("WITH: %w", err)
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
	err := q.IntoTable.AppendSQL(dialect, buf, args, params, env)
	if err != nil {
		return fmt.Errorf("INSERT INTO: %w", err)
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
		err = q.InsertColumns.AppendSQLExclude(dialect, buf, args, params, env, excludedTableQualifiers)
		if err != nil {
			return fmt.Errorf("INSERT INTO (columns): %w", err)
		}
		buf.WriteString(")")
	}
	// VALUES/SELECT
	switch {
	case len(q.RowValues) > 0:
		buf.WriteString(" VALUES ")
		err = q.RowValues.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("VALUES: %w", err)
		}
		if q.RowAlias != "" {
			if dialect != DialectMySQL {
				return fmt.Errorf("%s does not support row aliases", dialect)
			}
			buf.WriteString(" AS " + q.RowAlias)
			if len(q.ColumnAliases) > 0 {
				buf.WriteString(" (" + strings.Join(q.ColumnAliases, ", ") + ")")
			}
		}
	case q.SelectQuery != nil:
		buf.WriteString(" ")
		// sqlite INSERT ... SELECT FROM ... ON CONFLICT has parsing
		// ambiguity if the SELECT doesn't have a WHERE clause, so we
		// always make sure there is at least one predicate in the WHERE
		// clause
		if dialect == DialectSQLite && len(q.SelectQuery.WherePredicate.Predicates) == 0 {
			q.SelectQuery.WherePredicate.Predicates = []Predicate{Predicatef("TRUE")}
		}
		err = q.SelectQuery.AppendSQL(dialect, buf, args, params, env)
		if err != nil {
			return fmt.Errorf("SELECT: %w", err)
		}
	default:
		return fmt.Errorf("RowValues not provided and SelectQuery not provided to INSERT query")
	}
	// ON CONFLICT
	switch dialect {
	case DialectSQLite, DialectPostgres:
		if q.ConflictConstraint == "" && len(q.ConflictFields) == 0 && !q.ConflictDoNothing {
			break
		}
		buf.WriteString(" ON CONFLICT")
		if q.ConflictConstraint != "" {
			if dialect != DialectPostgres {
				return fmt.Errorf("%s does not support ON CONFLICT ON CONSTRAINT", dialect)
			}
			buf.WriteString(" ON CONSTRAINT " + q.ConflictConstraint)
		} else if len(q.ConflictFields) > 0 {
			buf.WriteString(" (")
			err = q.ConflictFields.AppendSQLExclude(dialect, buf, args, params, env, excludedTableQualifiers)
			if err != nil {
				return fmt.Errorf("ON CONFLICT (fields): %w", err)
			}
			buf.WriteString(")")
			if q.ConflictPredicate != nil {
				predicate := q.ConflictPredicate
				if p, ok := predicate.(VariadicPredicate); ok {
					p.Toplevel = true
					predicate = p
				}
				buf.WriteString(" WHERE ")
				err = predicate.AppendSQLExclude(dialect, buf, args, params, env, excludedTableQualifiers)
				if err != nil {
					return fmt.Errorf("ON CONFLICT ... WHERE: %w", err)
				}
			}
		}
		if len(q.Resolution) == 0 || q.ConflictDoNothing {
			buf.WriteString(" DO NOTHING")
			break
		}
		buf.WriteString(" DO UPDATE SET ")
		err = q.Resolution.AppendSQLExclude(dialect, buf, args, params, env, excludedTableQualifiers)
		if err != nil {
			return fmt.Errorf("DO UPDATE SET: %w", err)
		}
		if len(q.ResolutionPredicate.Predicates) > 0 {
			buf.WriteString(" WHERE ")
			q.ResolutionPredicate.Toplevel = true
			err = q.ResolutionPredicate.AppendSQLExclude(dialect, buf, args, params, env, nil)
			if err != nil {
				return fmt.Errorf("DO UPDATE SET ... WHERE: %w", err)
			}
		}
	case DialectMySQL:
		if len(q.Resolution) > 0 {
			buf.WriteString(" ON DUPLICATE KEY UPDATE ")
			err = q.Resolution.AppendSQLExclude(dialect, buf, args, params, nil, excludedTableQualifiers)
			if err != nil {
				return fmt.Errorf("ON DUPLICATE KEY UPDATE: %w", err)
			}
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

func (q InsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	switch q.Dialect {
	case DialectPostgres, DialectSQLite:
		q.ReturningFields = fields
		return q, nil
	default:
		return nil, fmt.Errorf("%s INSERT %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q InsertQuery) GetFetchableFields() ([]Field, error) {
	switch q.Dialect {
	case DialectPostgres, DialectSQLite:
		return q.ReturningFields, nil
	default:
		return nil, fmt.Errorf("%s INSERT %w", q.Dialect, ErrNonFetchableQuery)
	}
}

func (q InsertQuery) GetDialect() string { return q.Dialect }
