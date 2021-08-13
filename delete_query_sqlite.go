package sq

import "bytes"

type SQLiteDeleteQuery DeleteQuery

var _ Query = SQLiteDeleteQuery{}

func (q SQLiteDeleteQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return DeleteQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q SQLiteDeleteQuery) SetFetchableFields(fields []Field) (Query, error) {
	return DeleteQuery(q).SetFetchableFields(fields)
}

func (q SQLiteDeleteQuery) GetFetchableFields() ([]Field, error) {
	return DeleteQuery(q).GetFetchableFields()
}

func (q SQLiteDeleteQuery) GetDialect() string { return q.Dialect }

func (d SQLiteQueryBuilder) DeleteWith(ctes ...CTE) SQLiteDeleteQuery {
	return SQLiteDeleteQuery{
		Env:     d.env,
		Dialect: DialectSQLite,
		CTEs:    ctes,
	}
}

func (d SQLiteQueryBuilder) DeleteFrom(table SchemaTable) SQLiteDeleteQuery {
	return SQLiteDeleteQuery{
		Env:        d.env,
		Dialect:    DialectSQLite,
		FromTables: []SchemaTable{table},
	}
}

func (q SQLiteDeleteQuery) With(ctes ...CTE) SQLiteDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteDeleteQuery) DeleteFrom(table SchemaTable) SQLiteDeleteQuery {
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

func (q SQLiteDeleteQuery) OrderBy(fields ...Field) SQLiteDeleteQuery {
	q.OrderByFields = fields
	return q
}

func (q SQLiteDeleteQuery) Limit(limit int64) SQLiteDeleteQuery {
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}

func (q SQLiteDeleteQuery) Offset(offset int64) SQLiteDeleteQuery {
	q.RowOffset.Valid = true
	q.RowOffset.Int64 = offset
	return q
}

func (q SQLiteDeleteQuery) Returning(fields ...Field) SQLiteDeleteQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}
