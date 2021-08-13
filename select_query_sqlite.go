package sq

import "bytes"

type SQLiteSelectQuery SelectQuery

var _ Query = SQLiteSelectQuery{}

func (q SQLiteSelectQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return SelectQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q SQLiteSelectQuery) SetFetchableFields(fields []Field) (Query, error) {
	return SelectQuery(q).SetFetchableFields(fields)
}

func (q SQLiteSelectQuery) GetFetchableFields() ([]Field, error) {
	return SelectQuery(q).GetFetchableFields()
}

func (q SQLiteSelectQuery) GetDialect() string { return q.Dialect }

func (d SQLiteQueryBuilder) From(table Table) SQLiteSelectQuery {
	return SQLiteSelectQuery{
		Env:       d.env,
		Dialect:   DialectSQLite,
		FromTable: table,
	}
}

func (d SQLiteQueryBuilder) SelectWith(ctes ...CTE) SQLiteSelectQuery {
	return SQLiteSelectQuery{
		Env:     d.env,
		Dialect: DialectSQLite,
		CTEs:    ctes,
	}
}

func (d SQLiteQueryBuilder) Select(fields ...Field) SQLiteSelectQuery {
	return SQLiteSelectQuery{
		Env:          d.env,
		Dialect:      DialectSQLite,
		SelectFields: fields,
	}
}

func (d SQLiteQueryBuilder) SelectOne() SQLiteSelectQuery {
	return SQLiteSelectQuery{
		Env:          d.env,
		Dialect:      DialectSQLite,
		SelectFields: AliasFields{Literal("1")},
	}
}

func (d SQLiteQueryBuilder) SelectDistinct(fields ...Field) SQLiteSelectQuery {
	return SQLiteSelectQuery{
		Env:          d.env,
		Dialect:      DialectSQLite,
		Distinct:     true,
		SelectFields: fields,
	}
}

func (q SQLiteSelectQuery) With(ctes ...CTE) SQLiteSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteSelectQuery) Select(fields ...Field) SQLiteSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q SQLiteSelectQuery) SelectOne() SQLiteSelectQuery {
	q.SelectFields = AliasFields{Literal("1")}
	return q
}

func (q SQLiteSelectQuery) SelectDistinct(fields ...Field) SQLiteSelectQuery {
	q.Distinct = true
	q.SelectFields = fields
	return q
}

func (q SQLiteSelectQuery) From(table Table) SQLiteSelectQuery {
	q.FromTable = table
	return q
}

func (q SQLiteSelectQuery) Join(table Table, predicates ...Predicate) SQLiteSelectQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q SQLiteSelectQuery) LeftJoin(table Table, predicates ...Predicate) SQLiteSelectQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q SQLiteSelectQuery) CrossJoin(table Table) SQLiteSelectQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q SQLiteSelectQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) SQLiteSelectQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q SQLiteSelectQuery) Where(predicates ...Predicate) SQLiteSelectQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q SQLiteSelectQuery) GroupBy(fields ...Field) SQLiteSelectQuery {
	q.GroupByFields = append(q.GroupByFields, fields...)
	return q
}

func (q SQLiteSelectQuery) Having(predicates ...Predicate) SQLiteSelectQuery {
	q.HavingPredicate.Predicates = append(q.HavingPredicate.Predicates, predicates...)
	return q
}

func (q SQLiteSelectQuery) OrderBy(fields ...Field) SQLiteSelectQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q SQLiteSelectQuery) Limit(limit int64) SQLiteSelectQuery {
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}

func (q SQLiteSelectQuery) Offset(offset int64) SQLiteSelectQuery {
	q.RowOffset.Valid = true
	q.RowOffset.Int64 = offset
	return q
}
