package sq

type MySQLSelectQuery struct {
	SelectQuery
}

var _ Query = MySQLSelectQuery{}

func (d MySQLDialect) From(table Table) MySQLSelectQuery {
	var q MySQLSelectQuery
	q.QueryDialect = DialectMySQL
	q.FromTable = table
	return q
}

func (d MySQLDialect) Select(fields ...Field) MySQLSelectQuery {
	var q MySQLSelectQuery
	q.QueryDialect = DialectMySQL
	q.SelectFields = fields
	return q
}

func (d MySQLDialect) SelectWith(ctes ...CTE) MySQLSelectQuery {
	var q MySQLSelectQuery
	q.QueryDialect = DialectMySQL
	q.CTEs = ctes
	return q
}

func (d MySQLDialect) SelectDistinct(fields ...Field) MySQLSelectQuery {
	var q MySQLSelectQuery
	q.QueryDialect = DialectMySQL
	q.SelectType = SelectTypeDistinct
	q.SelectFields = fields
	return q
}

func (q MySQLSelectQuery) With(ctes ...CTE) MySQLSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q MySQLSelectQuery) Select(fields ...Field) MySQLSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q MySQLSelectQuery) SelectDistinct(fields ...Field) MySQLSelectQuery {
	q.SelectType = SelectTypeDistinct
	q.SelectFields = fields
	return q
}

func (q MySQLSelectQuery) From(table Table) MySQLSelectQuery {
	q.FromTable = table
	return q
}

func (q MySQLSelectQuery) Join(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q MySQLSelectQuery) LeftJoin(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q MySQLSelectQuery) RightJoin(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q MySQLSelectQuery) FullJoin(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q MySQLSelectQuery) CrossJoin(table Table) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q MySQLSelectQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q MySQLSelectQuery) Where(predicates ...Predicate) MySQLSelectQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q MySQLSelectQuery) GroupBy(fields ...Field) MySQLSelectQuery {
	q.GroupByFields = append(q.GroupByFields, fields...)
	return q
}

func (q MySQLSelectQuery) Having(predicates ...Predicate) MySQLSelectQuery {
	q.HavingPredicate.Predicates = append(q.HavingPredicate.Predicates, predicates...)
	return q
}

func (q MySQLSelectQuery) OrderBy(fields ...Field) MySQLSelectQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q MySQLSelectQuery) Limit(limit int64) MySQLSelectQuery {
	q.QueryLimit.Valid = true
	q.QueryLimit.Int64 = limit
	return q
}

func (q MySQLSelectQuery) Offset(offset int64) MySQLSelectQuery {
	q.QueryOffset.Valid = true
	q.QueryOffset.Int64 = offset
	return q
}