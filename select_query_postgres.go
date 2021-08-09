package sq

import "bytes"

type PostgresSelectQuery SelectQuery

var _ Query = PostgresSelectQuery{}

func (q PostgresSelectQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return SelectQuery(q).AppendSQL(dialect, buf, args, params)
}

func (q PostgresSelectQuery) SetFetchableFields(fields []Field) (Query, error) {
	return SelectQuery(q).SetFetchableFields(fields)
}

func (q PostgresSelectQuery) GetFetchableFields() ([]Field, error) {
	return SelectQuery(q).GetFetchableFields()
}

func (q PostgresSelectQuery) GetDialect() string { return q.Dialect }

func (d PostgresDialect) From(table Table) PostgresSelectQuery {
	var q PostgresSelectQuery
	q.Dialect = DialectPostgres
	q.FromTable = table
	return q
}

func (d PostgresDialect) SelectWith(ctes ...CTE) PostgresSelectQuery {
	var q PostgresSelectQuery
	q.Dialect = DialectPostgres
	q.CTEs = ctes
	return q
}

func (d PostgresDialect) Select(fields ...Field) PostgresSelectQuery {
	var q PostgresSelectQuery
	q.Dialect = DialectPostgres
	q.SelectFields = fields
	return q
}

func (d PostgresDialect) SelectOne() PostgresSelectQuery {
	var q PostgresSelectQuery
	q.Dialect = DialectPostgres
	q.SelectFields = AliasFields{Literal("1")}
	return q
}

func (d PostgresDialect) SelectDistinct(fields ...Field) PostgresSelectQuery {
	var q PostgresSelectQuery
	q.Dialect = DialectPostgres
	q.Distinct = true
	q.SelectFields = fields
	return q
}

func (q PostgresSelectQuery) With(ctes ...CTE) PostgresSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresSelectQuery) Select(fields ...Field) PostgresSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q PostgresSelectQuery) SelectDistinct(fields ...Field) PostgresSelectQuery {
	q.Distinct = true
	q.SelectFields = fields
	return q
}

func (q PostgresSelectQuery) DistinctOn(fields ...Field) PostgresSelectQuery {
	q.DistinctOnFields = fields
	return q
}

func (q PostgresSelectQuery) From(table Table) PostgresSelectQuery {
	q.FromTable = table
	return q
}

func (q PostgresSelectQuery) Join(table Table, predicates ...Predicate) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q PostgresSelectQuery) LeftJoin(table Table, predicates ...Predicate) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q PostgresSelectQuery) RightJoin(table Table, predicates ...Predicate) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q PostgresSelectQuery) FullJoin(table Table, predicates ...Predicate) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q PostgresSelectQuery) CrossJoin(table Table) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, CrossJoin(table))
	return q
}

func (q PostgresSelectQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q PostgresSelectQuery) Where(predicates ...Predicate) PostgresSelectQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q PostgresSelectQuery) GroupBy(fields ...Field) PostgresSelectQuery {
	q.GroupByFields = append(q.GroupByFields, fields...)
	return q
}

func (q PostgresSelectQuery) Having(predicates ...Predicate) PostgresSelectQuery {
	q.HavingPredicate.Predicates = append(q.HavingPredicate.Predicates, predicates...)
	return q
}

func (q PostgresSelectQuery) OrderBy(fields ...Field) PostgresSelectQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q PostgresSelectQuery) Limit(limit int64) PostgresSelectQuery {
	q.RowLimit.Valid = true
	q.RowLimit.Int64 = limit
	return q
}

func (q PostgresSelectQuery) Offset(offset int64) PostgresSelectQuery {
	q.RowOffset.Valid = true
	q.RowOffset.Int64 = offset
	return q
}
