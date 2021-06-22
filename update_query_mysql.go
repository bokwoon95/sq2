package sq

type MySQLUpdateQuery struct {
	UpdateQuery
}

var _ Query = MySQLUpdateQuery{}

func (d MySQLDialect) Update(table BaseTable) MySQLUpdateQuery {
	var q MySQLUpdateQuery
	q.QueryDialect = DialectMySQL
	q.UpdateTable = table
	return q
}

func (q MySQLUpdateQuery) Update(table BaseTable) MySQLUpdateQuery {
	q.UpdateTable = table
	return q
}

func (q MySQLUpdateQuery) Set(assignments ...Assignment) MySQLUpdateQuery {
	q.Assignments = append(q.Assignments, assignments...)
	return q
}

func (q MySQLUpdateQuery) Setx(mapper func(*Column) error) MySQLUpdateQuery {
	q.ColumnMapper = mapper
	return q
}

func (q MySQLUpdateQuery) From(table Table) MySQLUpdateQuery {
	q.FromTable = table
	return q
}

func (q MySQLUpdateQuery) Join(table Table, predicates ...Predicate) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q MySQLUpdateQuery) LeftJoin(table Table, predicates ...Predicate) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q MySQLUpdateQuery) RightJoin(table Table, predicates ...Predicate) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q MySQLUpdateQuery) CrossJoin(table Table) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q MySQLUpdateQuery) FullJoin(table Table, predicates ...Predicate) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q MySQLUpdateQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q MySQLUpdateQuery) Where(predicates ...Predicate) MySQLUpdateQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q MySQLUpdateQuery) OrderBy(fields ...Field) MySQLUpdateQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q MySQLUpdateQuery) Limit(limit int64) MySQLUpdateQuery {
	q.QueryLimit.Valid = true
	q.QueryLimit.Int64 = limit
	return q
}
