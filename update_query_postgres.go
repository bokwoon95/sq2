package sq

import "bytes"

type PostgresUpdateQuery UpdateQuery

var _ Query = PostgresUpdateQuery{}

func (q PostgresUpdateQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return UpdateQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q PostgresUpdateQuery) SetFetchableFields(fields []Field) (Query, error) {
	return UpdateQuery(q).SetFetchableFields(fields)
}

func (q PostgresUpdateQuery) GetFetchableFields() ([]Field, error) {
	return UpdateQuery(q).GetFetchableFields()
}

func (q PostgresUpdateQuery) GetDialect() string { return q.Dialect }

func (d PostgresQueryBuilder) UpdateWith(ctes ...CTE) PostgresUpdateQuery {
	return PostgresUpdateQuery{
		Env:     d.env,
		Dialect: DialectPostgres,
		CTEs:    ctes,
	}
}

func (d PostgresQueryBuilder) Update(table SchemaTable) PostgresUpdateQuery {
	return PostgresUpdateQuery{
		Env:         d.env,
		Dialect:     DialectPostgres,
		UpdateTable: table,
	}
}

func (q PostgresUpdateQuery) With(ctes ...CTE) PostgresUpdateQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresUpdateQuery) Update(table SchemaTable) PostgresUpdateQuery {
	q.UpdateTable = table
	return q
}

func (q PostgresUpdateQuery) Set(assignments ...Assignment) PostgresUpdateQuery {
	q.Assignments = append(q.Assignments, assignments...)
	return q
}

func (q PostgresUpdateQuery) Setx(mapper func(*Column) error) PostgresUpdateQuery {
	q.ColumnMapper = mapper
	return q
}

func (q PostgresUpdateQuery) From(table Table) PostgresUpdateQuery {
	q.FromTable = table
	return q
}

func (q PostgresUpdateQuery) Join(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) LeftJoin(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) RightJoin(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) FullJoin(table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q PostgresUpdateQuery) CrossJoin(table Table) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q PostgresUpdateQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) PostgresUpdateQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q PostgresUpdateQuery) Where(predicates ...Predicate) PostgresUpdateQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q PostgresUpdateQuery) Returning(fields ...Field) PostgresUpdateQuery {
	q.ReturningFields = append(q.ReturningFields, fields...)
	return q
}
