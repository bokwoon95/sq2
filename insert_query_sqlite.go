package sq

import "bytes"

type SQLiteInsertQuery InsertQuery

var _ Query = SQLiteInsertQuery{}

func (q SQLiteInsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return InsertQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q SQLiteInsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	return InsertQuery(q).SetFetchableFields(fields)
}

func (q SQLiteInsertQuery) GetFetchableFields() ([]Field, error) {
	return InsertQuery(q).GetFetchableFields()
}

func (q SQLiteInsertQuery) GetDialect() string { return q.Dialect }

func (d SQLiteQueryBuilder) InsertWith(ctes ...CTE) SQLiteInsertQuery {
	return SQLiteInsertQuery{
		Env:     d.env,
		Dialect: DialectSQLite,
		CTEs:    ctes,
	}
}

func (d SQLiteQueryBuilder) InsertInto(table SchemaTable) SQLiteInsertQuery {
	return SQLiteInsertQuery{
		Env:       d.env,
		Dialect:   DialectSQLite,
		IntoTable: table,
	}
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
	selectQuery := SelectQuery(query)
	q.SelectQuery = &selectQuery
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

func (c SQLiteInsertConflict) Where(conflictPredicate Predicate) SQLiteInsertConflict {
	c.insertQuery.ConflictPredicate = conflictPredicate
	return c
}

func (c SQLiteInsertConflict) DoNothing() SQLiteInsertQuery {
	q := c.insertQuery
	q.ConflictDoNothing = true
	return *q
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
