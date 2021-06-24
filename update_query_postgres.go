package sq

type PostgresUpdateQuery struct {
	UpdateQuery
}

var _ Query = PostgresUpdateQuery{}

func (d PostgresDialect) UpdateWith(ctes ...CTE) PostgresUpdateQuery {
	var q PostgresUpdateQuery
	q.QueryDialect = DialectPostgres
	q.CTEs = ctes
	return q
}

func (d PostgresDialect) Update(table BaseTable) PostgresUpdateQuery {
	var q PostgresUpdateQuery
	q.QueryDialect = DialectPostgres
	q.UpdateTable = table
	return q
}

func (q PostgresUpdateQuery) With(ctes ...CTE) PostgresUpdateQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresUpdateQuery) Update(table BaseTable) PostgresUpdateQuery {
	q.UpdateTable = table
	return q
}

func (q PostgresUpdateQuery) Set(assignments ...Assignment) PostgresUpdateQuery {
	q.Assignments = append(q.Assignments, assignments...)
	return q
}

func (q PostgresUpdateQuery) Setx(mapper func(*Column) error) PostgresUpdateQuery {
	q.ColumnMapper = mapper
	return q
}

func (q PostgresUpdateQuery) From(table Table) PostgresUpdateQuery {
	q.FromTable = table
	return q
}

func (q PostgresUpdateQuery) Join(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) LeftJoin(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) RightJoin(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) CrossJoin(table Table) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q PostgresUpdateQuery) FullJoin(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q PostgresUpdateQuery) Where(predicates ...Predicate) PostgresUpdateQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q PostgresUpdateQuery) Returning(fields ...Field) PostgresUpdateQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}