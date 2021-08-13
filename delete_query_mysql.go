package sq

import "bytes"

type MySQLDeleteQuery DeleteQuery

var _ Query = MySQLDeleteQuery{}

func (q MySQLDeleteQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return DeleteQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q MySQLDeleteQuery) SetFetchableFields(fields []Field) (Query, error) {
	return DeleteQuery(q).SetFetchableFields(fields)
}

func (q MySQLDeleteQuery) GetFetchableFields() ([]Field, error) {
	return DeleteQuery(q).GetFetchableFields()
}

func (q MySQLDeleteQuery) GetDialect() string { return q.Dialect }

func (d MySQLQueryBuilder) DeleteWith(ctes ...CTE) MySQLDeleteQuery {
	return MySQLDeleteQuery{
		Env:     d.env,
		Dialect: DialectMySQL,
		CTEs:    ctes,
	}
}

func (d MySQLQueryBuilder) DeleteFrom(tables ...SchemaTable) MySQLDeleteQuery {
	return MySQLDeleteQuery{
		Env:        d.env,
		Dialect:    DialectMySQL,
		FromTables: tables,
	}
}

func (q MySQLDeleteQuery) With(ctes ...CTE) MySQLDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q MySQLDeleteQuery) DeleteFrom(tables ...SchemaTable) MySQLDeleteQuery {
	q.FromTables = tables
	return q
}

func (q MySQLDeleteQuery) Using(table Table) MySQLDeleteQuery {
	q.UsingTable = table
	return q
}

func (q MySQLDeleteQuery) Join(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) LeftJoin(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) RightJoin(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) FullJoin(table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q MySQLDeleteQuery) CrossJoin(table Table) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q MySQLDeleteQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) MySQLDeleteQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q MySQLDeleteQuery) Where(predicates ...Predicate) MySQLDeleteQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q MySQLDeleteQuery) OrderBy(fields ...Field) MySQLDeleteQuery {
	q.OrderByFields = fields
	return q
}

func (q MySQLDeleteQuery) Limit(limit int64) MySQLDeleteQuery {
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}
