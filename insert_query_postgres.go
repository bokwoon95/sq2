package sq

import "bytes"

type PostgresInsertQuery InsertQuery

var _ Query = PostgresInsertQuery{}

func (q PostgresInsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return InsertQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q PostgresInsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	return InsertQuery(q).SetFetchableFields(fields)
}

func (q PostgresInsertQuery) GetFetchableFields() ([]Field, error) {
	return InsertQuery(q).GetFetchableFields()
}

func (q PostgresInsertQuery) GetDialect() string { return q.Dialect }

func (d PostgresQueryBuilder) InsertWith(ctes ...CTE) PostgresInsertQuery {
	return PostgresInsertQuery{
		Env:     d.env,
		Dialect: DialectPostgres,
		CTEs:    ctes,
	}
}

func (d PostgresQueryBuilder) InsertInto(table SchemaTable) PostgresInsertQuery {
	return PostgresInsertQuery{
		Env:       d.env,
		Dialect:   DialectPostgres,
		IntoTable: table,
	}
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
	selectQuery := SelectQuery(query)
	q.SelectQuery = &selectQuery
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

func (q PostgresInsertQuery) OnConflictOnConstraint(constraintName string) PostgresInsertConflict {
	var c PostgresInsertConflict
	q.ConflictConstraint = constraintName
	c.insertQuery = &q
	return c
}

func (c PostgresInsertConflict) Where(conflictPredicate Predicate) PostgresInsertConflict {
	c.insertQuery.ConflictPredicate = conflictPredicate
	return c
}

func (c PostgresInsertConflict) DoNothing() PostgresInsertQuery {
	q := c.insertQuery
	q.ConflictDoNothing = true
	return *q
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
