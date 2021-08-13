package sq

import "bytes"

type MySQLSelectQuery SelectQuery

var _ Query = MySQLSelectQuery{}

func (q MySQLSelectQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return SelectQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q MySQLSelectQuery) SetFetchableFields(fields []Field) (Query, error) {
	return SelectQuery(q).SetFetchableFields(fields)
}

func (q MySQLSelectQuery) GetFetchableFields() ([]Field, error) {
	return SelectQuery(q).GetFetchableFields()
}

func (q MySQLSelectQuery) GetDialect() string { return q.Dialect }

func (d MySQLQueryBuilder) From(table Table) MySQLSelectQuery {
	return MySQLSelectQuery{
		Env:       d.env,
		Dialect:   DialectMySQL,
		FromTable: table,
	}
}

func (d MySQLQueryBuilder) SelectWith(ctes ...CTE) MySQLSelectQuery {
	return MySQLSelectQuery{
		Env:     d.env,
		Dialect: DialectMySQL,
		CTEs:    ctes,
	}
}

func (d MySQLQueryBuilder) Select(fields ...Field) MySQLSelectQuery {
	return MySQLSelectQuery{
		Env:          d.env,
		Dialect:      DialectMySQL,
		SelectFields: fields,
	}
}

func (d MySQLQueryBuilder) SelectOne() MySQLSelectQuery {
	return MySQLSelectQuery{
		Env:          d.env,
		Dialect:      DialectMySQL,
		SelectFields: AliasFields{Literal("1")},
	}
}

func (d MySQLQueryBuilder) SelectDistinct(fields ...Field) MySQLSelectQuery {
	return MySQLSelectQuery{
		Env:          d.env,
		Dialect:      DialectMySQL,
		Distinct:     true,
		SelectFields: fields,
	}
}

func (q MySQLSelectQuery) With(ctes ...CTE) MySQLSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q MySQLSelectQuery) Select(fields ...Field) MySQLSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q MySQLSelectQuery) SelectDistinct(fields ...Field) MySQLSelectQuery {
	q.Distinct = true
	q.SelectFields = fields
	return q
}

func (q MySQLSelectQuery) From(table Table) MySQLSelectQuery {
	q.FromTable = table
	return q
}

func (q MySQLSelectQuery) Join(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q MySQLSelectQuery) LeftJoin(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q MySQLSelectQuery) RightJoin(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q MySQLSelectQuery) FullJoin(table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q MySQLSelectQuery) CrossJoin(table Table) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q MySQLSelectQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) MySQLSelectQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q MySQLSelectQuery) Where(predicates ...Predicate) MySQLSelectQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q MySQLSelectQuery) GroupBy(fields ...Field) MySQLSelectQuery {
	q.GroupByFields = append(q.GroupByFields, fields...)
	return q
}

func (q MySQLSelectQuery) Having(predicates ...Predicate) MySQLSelectQuery {
	q.HavingPredicate.Predicates = append(q.HavingPredicate.Predicates, predicates...)
	return q
}

func (q MySQLSelectQuery) OrderBy(fields ...Field) MySQLSelectQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q MySQLSelectQuery) Limit(limit int64) MySQLSelectQuery {
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}

func (q MySQLSelectQuery) Offset(offset int64) MySQLSelectQuery {
	q.RowOffset.Valid = true
	q.RowOffset.Int64 = offset
	return q
}
