package sq

type MySQLInsertQuery struct {
	InsertQuery
}

var _ Query = MySQLInsertQuery{}

func (d MySQLDialect) InsertWith(ctes ...CTE) MySQLInsertQuery {
	var q MySQLInsertQuery
	q.QueryDialect = DialectMySQL
	q.CTEs = ctes
	return q
}

func (d MySQLDialect) InsertInto(table BaseTable) MySQLInsertQuery {
	var q MySQLInsertQuery
	q.QueryDialect = DialectMySQL
	q.IntoTable = table
	return q
}

func (q MySQLInsertQuery) With(ctes ...CTE) MySQLInsertQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q MySQLInsertQuery) InsertInto(table BaseTable) MySQLInsertQuery {
	q.IntoTable = table
	return q
}

func (q MySQLInsertQuery) Columns(fields ...Field) MySQLInsertQuery {
	q.InsertColumns = fields
	return q
}

func (q MySQLInsertQuery) Values(values ...interface{}) MySQLInsertQuery {
	q.RowValues = append(q.RowValues, values)
	return q
}

func (q MySQLInsertQuery) Valuesx(mapper func(*Column) error) MySQLInsertQuery {
	q.ColumnMapper = mapper
	return q
}

func (q MySQLInsertQuery) Select(query MySQLSelectQuery) MySQLInsertQuery {
	q.SelectQuery = &query.SelectQuery
	return q
}

func (q MySQLInsertQuery) OnDuplicateKeyUpdate(assignments ...Assignment) MySQLInsertQuery {
	q.Resolution = assignments
	return q
}

func (q MySQLInsertQuery) Where(predicates ...Predicate) MySQLInsertQuery {
	q.ResolutionPredicate.Predicates = append(q.ResolutionPredicate.Predicates, predicates...)
	return q
}
