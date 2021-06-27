package sq

type PostgresDeleteQuery struct {
	DeleteQuery
}

var _ Query = PostgresDeleteQuery{}

func (d PostgresDialect) DeleteWith(ctes ...CTE) PostgresDeleteQuery {
	var q PostgresDeleteQuery
	q.QueryDialect = DialectPostgres
	q.CTEs = ctes
	return q
}

func (d PostgresDialect) DeleteFrom(table BaseTable) PostgresDeleteQuery {
	var q PostgresDeleteQuery
	q.QueryDialect = DialectPostgres
	q.FromTables = append(q.FromTables, table)
	return q
}

func (q PostgresDeleteQuery) With(ctes ...CTE) PostgresDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresDeleteQuery) DeleteFrom(table BaseTable) PostgresDeleteQuery {
	if len(q.FromTables) == 0 {
		q.FromTables = append(q.FromTables, table)
	} else {
		q.FromTables[0] = table
		q.FromTables = q.FromTables[:1]
	}
	return q
}

func (q PostgresDeleteQuery) Using(table Table) PostgresDeleteQuery {
	q.UsingTable = table
	return q
}

func (q PostgresDeleteQuery) Join(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) LeftJoin(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) RightJoin(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) FullJoin(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) CrossJoin(table Table) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q PostgresDeleteQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q PostgresDeleteQuery) Where(predicates ...Predicate) PostgresDeleteQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q PostgresDeleteQuery) Returning(fields ...Field) PostgresDeleteQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}
