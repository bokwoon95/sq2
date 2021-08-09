package sq

import "bytes"

type PostgresInsertQuery InsertQuery

var _ Query = PostgresInsertQuery{}

func (q PostgresInsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return InsertQuery(q).AppendSQL(dialect, buf, args, params)
}

func (q PostgresInsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	return InsertQuery(q).SetFetchableFields(fields)
}

func (q PostgresInsertQuery) GetFetchableFields() ([]Field, error) {
	return InsertQuery(q).GetFetchableFields()
}

func (q PostgresInsertQuery) GetDialect() string { return q.Dialect }

func (d PostgresDialect) InsertWith(ctes ...CTE) PostgresInsertQuery {
	var q PostgresInsertQuery
	q.Dialect = DialectPostgres
	q.CTEs = ctes
	return q
}

func (d PostgresDialect) InsertInto(table SchemaTable) PostgresInsertQuery {
	var q PostgresInsertQuery
	q.Dialect = DialectPostgres
	q.IntoTable = table
	return q
}

func (q PostgresInsertQuery) With(ctes ...CTE) PostgresInsertQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresInsertQuery) InsertInto(table SchemaTable) PostgresInsertQuery {
	q.IntoTable = table
	return q
}

func (q PostgresInsertQuery) Columns(fields ...Field) PostgresInsertQuery {
	q.InsertColumns = fields
	return q
}

func (q PostgresInsertQuery) Values(values ...interface{}) PostgresInsertQuery {
	q.RowValues = append(q.RowValues, values)
	return q
}

func (q PostgresInsertQuery) Valuesx(mapper func(*Column) error) PostgresInsertQuery {
	q.ColumnMapper = mapper
	return q
}

func (q PostgresInsertQuery) Select(query PostgresSelectQuery) PostgresInsertQuery {
	q.SelectQuery = &query.SelectQuery
	return q
}

type PostgresInsertConflict struct {
	insertQuery *PostgresInsertQuery
}

func (q PostgresInsertQuery) OnConflict(fields ...Field) PostgresInsertConflict {
	var c PostgresInsertConflict
	q.ConflictFields = fields
	c.insertQuery = &q
	return c
}

func (q PostgresInsertQuery) OnConflictOnConstraint(name string) PostgresInsertConflict {
	var c PostgresInsertConflict
	q.ConflictConstraint = name
	c.insertQuery = &q
	return c
}

func (c PostgresInsertConflict) Where(predicates ...Predicate) PostgresInsertConflict {
	c.insertQuery.ConflictPredicate.Predicates = append(c.insertQuery.ConflictPredicate.Predicates, predicates...)
	return c
}

func (c PostgresInsertConflict) DoNothing() PostgresInsertQuery {
	return *c.insertQuery
}

func (c PostgresInsertConflict) DoUpdateSet(assignments ...Assignment) PostgresInsertQuery {
	c.insertQuery.Resolution = assignments
	return *c.insertQuery
}

func (q PostgresInsertQuery) Where(predicates ...Predicate) PostgresInsertQuery {
	q.ResolutionPredicate.Predicates = append(q.ResolutionPredicate.Predicates, predicates...)
	return q
}

func (q PostgresInsertQuery) Returning(fields ...Field) PostgresInsertQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}
