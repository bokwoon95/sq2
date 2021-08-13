package sq

import "bytes"

type MySQLUpdateQuery UpdateQuery

var _ Query = MySQLUpdateQuery{}

func (q MySQLUpdateQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return UpdateQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q MySQLUpdateQuery) SetFetchableFields(fields []Field) (Query, error) {
	return UpdateQuery(q).SetFetchableFields(fields)
}

func (q MySQLUpdateQuery) GetFetchableFields() ([]Field, error) {
	return UpdateQuery(q).GetFetchableFields()
}

func (q MySQLUpdateQuery) GetDialect() string { return q.Dialect }

func (d MySQLQueryBuilder) UpdateWith(ctes ...CTE) MySQLUpdateQuery {
	return MySQLUpdateQuery{
		Env:     d.env,
		Dialect: DialectMySQL,
		CTEs:    ctes,
	}
}

func (d MySQLQueryBuilder) Update(table SchemaTable) MySQLUpdateQuery {
	return MySQLUpdateQuery{
		Env:         d.env,
		Dialect:     DialectMySQL,
		UpdateTable: table,
	}
}

func (q MySQLUpdateQuery) With(ctes ...CTE) MySQLUpdateQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q MySQLUpdateQuery) Update(table SchemaTable) MySQLUpdateQuery {
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

func (q MySQLUpdateQuery) FullJoin(table Table, predicates ...Predicate) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q MySQLUpdateQuery) CrossJoin(table Table) MySQLUpdateQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
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
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}
