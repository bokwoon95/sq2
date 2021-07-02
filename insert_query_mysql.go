package sq

type MySQLInsertQuery struct {
	InsertQuery
}

var _ Query = MySQLInsertQuery{}

func (d MySQLDialect) InsertInto(table SchemaTable) MySQLInsertQuery {
	var q MySQLInsertQuery
	q.Dialect = DialectMySQL
	q.IntoTable = table
	return q
}

func (d MySQLDialect) InsertIgnoreInto(table SchemaTable) MySQLInsertQuery {
	var q MySQLInsertQuery
	q.Dialect = DialectMySQL
	q.InsertIgnore = true
	q.IntoTable = table
	return q
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

func (q MySQLInsertQuery) As(rowAlias string, columnAliases ...string) MySQLInsertQuery {
	q.RowAlias = rowAlias
	q.ColumnAliases = columnAliases
	return q
}

func (q MySQLInsertQuery) Select(query MySQLSelectQuery) MySQLInsertQuery {
	q.SelectQuery = &query.SelectQuery
	return q
}

func (q MySQLInsertQuery) OnDuplicateKeyUpdate(assignments ...Assignment) MySQLInsertQuery {
	q.Resolution = assignments
	return q
}
