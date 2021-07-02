package sq

type SQLiteInsertQuery struct {
	InsertQuery
}

var _ Query = SQLiteInsertQuery{}

func (d SQLiteDialect) InsertWith(ctes ...CTE) SQLiteInsertQuery {
	var q SQLiteInsertQuery
	q.Dialect = DialectSQLite
	q.CTEs = ctes
	return q
}

func (d SQLiteDialect) InsertInto(table SchemaTable) SQLiteInsertQuery {
	var q SQLiteInsertQuery
	q.Dialect = DialectSQLite
	q.IntoTable = table
	return q
}

func (q SQLiteInsertQuery) With(ctes ...CTE) SQLiteInsertQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteInsertQuery) InsertInto(table SchemaTable) SQLiteInsertQuery {
	q.IntoTable = table
	return q
}

func (q SQLiteInsertQuery) Columns(fields ...Field) SQLiteInsertQuery {
	q.InsertColumns = fields
	return q
}

func (q SQLiteInsertQuery) Values(values ...interface{}) SQLiteInsertQuery {
	q.RowValues = append(q.RowValues, values)
	return q
}

func (q SQLiteInsertQuery) Valuesx(mapper func(*Column) error) SQLiteInsertQuery {
	q.ColumnMapper = mapper
	return q
}

func (q SQLiteInsertQuery) Select(query SQLiteSelectQuery) SQLiteInsertQuery {
	q.SelectQuery = &query.SelectQuery
	return q
}

type SQLiteInsertConflict struct {
	insertQuery *SQLiteInsertQuery
}

func (q SQLiteInsertQuery) OnConflict(fields ...Field) SQLiteInsertConflict {
	var c SQLiteInsertConflict
	q.ConflictFields = fields
	c.insertQuery = &q
	return c
}

func (c SQLiteInsertConflict) Where(predicates ...Predicate) SQLiteInsertConflict {
	c.insertQuery.ConflictPredicate.Predicates = append(c.insertQuery.ConflictPredicate.Predicates, predicates...)
	return c
}

func (c SQLiteInsertConflict) DoNothing() SQLiteInsertQuery {
	return *c.insertQuery
}

func (c SQLiteInsertConflict) DoUpdateSet(assignments ...Assignment) SQLiteInsertQuery {
	c.insertQuery.Resolution = assignments
	return *c.insertQuery
}

func (q SQLiteInsertQuery) Where(predicates ...Predicate) SQLiteInsertQuery {
	q.ResolutionPredicate.Predicates = append(q.ResolutionPredicate.Predicates, predicates...)
	return q
}

func (q SQLiteInsertQuery) Returning(fields ...Field) SQLiteInsertQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}
