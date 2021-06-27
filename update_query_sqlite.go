package sq

type SQLiteUpdateQuery struct {
	UpdateQuery
}

var _ Query = SQLiteUpdateQuery{}

func (d SQLiteDialect) UpdateWith(ctes ...CTE) SQLiteUpdateQuery {
	var q SQLiteUpdateQuery
	q.Dialect = DialectSQLite
	q.CTEs = ctes
	return q
}

func (d SQLiteDialect) Update(table BaseTable) SQLiteUpdateQuery {
	var q SQLiteUpdateQuery
	q.Dialect = DialectSQLite
	q.UpdateTable = table
	return q
}

func (q SQLiteUpdateQuery) With(ctes ...CTE) SQLiteUpdateQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteUpdateQuery) Update(table BaseTable) SQLiteUpdateQuery {
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

func (q SQLiteUpdateQuery) RightJoin(table Table, predicates ...Predicate) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) CrossJoin(table Table) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q SQLiteUpdateQuery) FullJoin(table Table, predicates ...Predicate) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
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
