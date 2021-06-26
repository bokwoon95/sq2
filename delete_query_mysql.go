package sq

type MySQLDeleteQuery struct {
	DeleteQuery
}

var _ Query = MySQLDeleteQuery{}

func (d MySQLDialect) DeleteWith(ctes ...CTE) MySQLDeleteQuery {
	var q MySQLDeleteQuery
	q.QueryDialect = DialectMySQL
	q.CTEs = ctes
	return q
}

func (d MySQLDialect) DeleteFrom(table BaseTable) MySQLDeleteQuery {
	var q MySQLDeleteQuery
	q.QueryDialect = DialectMySQL
	q.FromTable = table
	return q
}

func (q MySQLDeleteQuery) With(ctes ...CTE) MySQLDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q MySQLDeleteQuery) DeleteFrom(table BaseTable) MySQLDeleteQuery {
	q.FromTable = table
	return q
}

func (q MySQLDeleteQuery) Using(table Table) MySQLDeleteQuery {
	q.UsingTable = table
	return q
}

func (q MySQLDeleteQuery) Join(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) LeftJoin(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) RightJoin(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) FullJoin(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) CrossJoin(table Table) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q MySQLDeleteQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q MySQLDeleteQuery) Where(predicates ...Predicate) MySQLDeleteQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}
