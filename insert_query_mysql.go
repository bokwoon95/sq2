package sq

import "bytes"

type MySQLInsertQuery InsertQuery

var _ Query = MySQLInsertQuery{}

func (q MySQLInsertQuery) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int, env map[string]interface{}) error {
	return InsertQuery(q).AppendSQL(dialect, buf, args, params, env)
}

func (q MySQLInsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	return InsertQuery(q).SetFetchableFields(fields)
}

func (q MySQLInsertQuery) GetFetchableFields() ([]Field, error) {
	return InsertQuery(q).GetFetchableFields()
}

func (q MySQLInsertQuery) GetDialect() string { return q.Dialect }

func (d MySQLQueryBuilder) InsertInto(table SchemaTable) MySQLInsertQuery {
	return MySQLInsertQuery{
		Env:       d.env,
		Dialect:   DialectMySQL,
		IntoTable: table,
	}
}

func (d MySQLQueryBuilder) InsertIgnoreInto(table SchemaTable) MySQLInsertQuery {
	return MySQLInsertQuery{
		Env:          d.env,
		Dialect:      DialectMySQL,
		InsertIgnore: true,
		IntoTable:    table,
	}
}

func (q MySQLInsertQuery) Columns(fields ...Field) MySQLInsertQuery {
	q.InsertColumns = fields
	return q
}

func (q MySQLInsertQuery) Values(values ...interface{}) MySQLInsertQuery {
	q.RowValues = append(q.RowValues, values)
	return q
}

func (q MySQLInsertQuery) Valuesx(mapper func(*Column) error) MySQLInsertQuery {
	q.ColumnMapper = mapper
	return q
}

func (q MySQLInsertQuery) AsRow(rowAlias string) MySQLInsertQuery {
	q.RowAlias = rowAlias
	return q
}

func (q MySQLInsertQuery) AsColumns(columnAliases ...string) MySQLInsertQuery {
	q.ColumnAliases = columnAliases
	return q
}

func (q MySQLInsertQuery) Select(query MySQLSelectQuery) MySQLInsertQuery {
	selectQuery := SelectQuery(query)
	q.SelectQuery = &selectQuery
	return q
}

func (q MySQLInsertQuery) OnDuplicateKeyUpdate(assignments ...Assignment) MySQLInsertQuery {
	q.Resolution = assignments
	return q
}
