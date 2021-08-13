package sq

import "bytes"

type SQLiteUpdateQuery UpdateQuery

var _ Query = SQLiteUpdateQuery{}

func (q SQLiteUpdateQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return UpdateQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q SQLiteUpdateQuery) SetFetchableFields(fields []Field) (Query, error) {
	return UpdateQuery(q).SetFetchableFields(fields)
}

func (q SQLiteUpdateQuery) GetFetchableFields() ([]Field, error) {
	return UpdateQuery(q).GetFetchableFields()
}

func (q SQLiteUpdateQuery) GetDialect() string { return q.Dialect }

func (d SQLiteQueryBuilder) UpdateWith(ctes ...CTE) SQLiteUpdateQuery {
	return SQLiteUpdateQuery{
		Env:     d.env,
		Dialect: DialectSQLite,
		CTEs:    ctes,
	}
}

func (d SQLiteQueryBuilder) Update(table SchemaTable) SQLiteUpdateQuery {
	return SQLiteUpdateQuery{
		Env:         d.env,
		Dialect:     DialectSQLite,
		UpdateTable: table,
	}
}

func (q SQLiteUpdateQuery) With(ctes ...CTE) SQLiteUpdateQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteUpdateQuery) Update(table SchemaTable) SQLiteUpdateQuery {
	q.UpdateTable = table
	return q
}

func (q SQLiteUpdateQuery) Set(assignments ...Assignment) SQLiteUpdateQuery {
	q.Assignments = append(q.Assignments, assignments...)
	return q
}

func (q SQLiteUpdateQuery) Setx(mapper func(*Column) error) SQLiteUpdateQuery {
	q.ColumnMapper = mapper
	return q
}

func (q SQLiteUpdateQuery) From(table Table) SQLiteUpdateQuery {
	q.FromTable = table
	return q
}

func (q SQLiteUpdateQuery) Join(table Table, predicates ...Predicate) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) LeftJoin(table Table, predicates ...Predicate) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) CrossJoin(table Table) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q SQLiteUpdateQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) Where(predicates ...Predicate) SQLiteUpdateQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q SQLiteUpdateQuery) OrderBy(fields ...Field) SQLiteUpdateQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q SQLiteUpdateQuery) Limit(limit int64) SQLiteUpdateQuery {
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}

func (q SQLiteUpdateQuery) Offset(offset int64) SQLiteUpdateQuery {
	q.RowOffset.Valid = true
	q.RowOffset.Int64 = offset
	return q
}
