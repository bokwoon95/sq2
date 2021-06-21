package sq

import "bytes"

type InsertQuery struct {
	QueryDialect string
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// INSERT INTO
	Ignore        bool
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

func (q InsertQuery) ToSQL() (query string, args []interface{}, params map[string][]int, err error) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string][]int)
	err = q.AppendSQL(q.QueryDialect, buf, &args, params)
	return buf.String(), args, params, err
}

// TODO: change `excludedTableQualifiers []string` to `excludeTableQualifier func(name string) bool`

func (q InsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	var err error
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
		err = q.CTEs.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
	}
	// INSERT INTO
	buf.WriteString("INSERT INTO ")
	if q.IntoTable == nil {
		buf.WriteString("NULL")
	} else {
		err = q.IntoTable.AppendSQL(dialect, buf, args, params)
		if err != nil {
			return err
		}
		name := q.IntoTable.GetName()
		alias := q.IntoTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(alias)
			excludedTableQualifiers = append(excludedTableQualifiers, alias)
		} else {
			excludedTableQualifiers = append(excludedTableQualifiers, name)
		}
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
	}
	// ON CONFLICT
	if q.HandleConflict {
		buf.WriteString(" ON CONFLICT")
		if len(q.ConflictFields) > 0 {
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
		}
		if len(q.Resolution) > 0 {
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
		} else {
			buf.WriteString(" DO NOTHING")
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
