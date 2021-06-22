package sq

type SQLiteSelectQuery struct {
	SelectQuery
}

var _ Query = SQLiteSelectQuery{}

func (d SQLiteDialect) From(table Table) SQLiteSelectQuery {
	var q SQLiteSelectQuery
	q.QueryDialect = DialectSQLite
	q.FromTable = table
	return q
}

func (d SQLiteDialect) SelectWith(ctes ...CTE) SQLiteSelectQuery {
	var q SQLiteSelectQuery
	q.QueryDialect = DialectSQLite
	q.CTEs = ctes
	return q
}

func (d SQLiteDialect) Select(fields ...Field) SQLiteSelectQuery {
	var q SQLiteSelectQuery
	q.QueryDialect = DialectSQLite
	q.SelectFields = fields
	return q
}

func (d SQLiteDialect) SelectDistinct(fields ...Field) SQLiteSelectQuery {
	var q SQLiteSelectQuery
	q.QueryDialect = DialectSQLite
	q.SelectType = SelectTypeDistinct
	q.SelectFields = fields
	return q
}

func (q SQLiteSelectQuery) With(ctes ...CTE) SQLiteSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteSelectQuery) Select(fields ...Field) SQLiteSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q SQLiteSelectQuery) SelectDistinct(fields ...Field) SQLiteSelectQuery {
	q.SelectType = SelectTypeDistinct
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
	q.QueryLimit.Valid = true
	q.QueryLimit.Int64 = limit
	return q
}

func (q SQLiteSelectQuery) Offset(offset int64) SQLiteSelectQuery {
	q.QueryOffset.Valid = true
	q.QueryOffset.Int64 = offset
	return q
}
