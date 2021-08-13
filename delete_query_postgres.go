package sq

import "bytes"

type PostgresDeleteQuery DeleteQuery

var _ Query = PostgresDeleteQuery{}

func (q PostgresDeleteQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return DeleteQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q PostgresDeleteQuery) SetFetchableFields(fields []Field) (Query, error) {
	return DeleteQuery(q).SetFetchableFields(fields)
}

func (q PostgresDeleteQuery) GetFetchableFields() ([]Field, error) {
	return DeleteQuery(q).GetFetchableFields()
}

func (q PostgresDeleteQuery) GetDialect() string { return q.Dialect }

func (d PostgresQueryBuilder) DeleteWith(ctes ...CTE) PostgresDeleteQuery {
	return PostgresDeleteQuery{
		Env:     d.env,
		Dialect: DialectPostgres,
		CTEs:    ctes,
	}
}

func (d PostgresQueryBuilder) DeleteFrom(table SchemaTable) PostgresDeleteQuery {
	return PostgresDeleteQuery{
		Env:        d.env,
		Dialect:    DialectPostgres,
		FromTables: []SchemaTable{table},
	}
}

func (q PostgresDeleteQuery) With(ctes ...CTE) PostgresDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresDeleteQuery) DeleteFrom(table SchemaTable) PostgresDeleteQuery {
	if len(q.FromTables) == 0 {
		q.FromTables = append(q.FromTables, table)
	} else {
		q.FromTables[0] = table
		q.FromTables = q.FromTables[:1]
	}
	return q
}

func (q PostgresDeleteQuery) Using(table Table) PostgresDeleteQuery {
	q.UsingTable = table
	return q
}

func (q PostgresDeleteQuery) Join(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) LeftJoin(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) RightJoin(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) FullJoin(table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q PostgresDeleteQuery) CrossJoin(table Table) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q PostgresDeleteQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) PostgresDeleteQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q PostgresDeleteQuery) Where(predicates ...Predicate) PostgresDeleteQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q PostgresDeleteQuery) Returning(fields ...Field) PostgresDeleteQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}
