package sq

type SQLiteDeleteQuery struct {
	DeleteQuery
}

var _ Query = SQLiteDeleteQuery{}

func (d SQLiteDialect) DeleteWith(ctes ...CTE) SQLiteDeleteQuery {
	var q SQLiteDeleteQuery
	q.Dialect = DialectSQLite
	q.CTEs = ctes
	return q
}

func (d SQLiteDialect) DeleteFrom(table BaseTable) SQLiteDeleteQuery {
	var q SQLiteDeleteQuery
	q.Dialect = DialectSQLite
	q.FromTables = append(q.FromTables, table)
	return q
}

func (q SQLiteDeleteQuery) With(ctes ...CTE) SQLiteDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteDeleteQuery) DeleteFrom(table BaseTable) SQLiteDeleteQuery {
	if len(q.FromTables) == 0 {
		q.FromTables = append(q.FromTables, table)
	} else {
		q.FromTables[0] = table
		q.FromTables = q.FromTables[:1]
	}
	return q
}

func (q SQLiteDeleteQuery) Where(predicates ...Predicate) SQLiteDeleteQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}
