package ddl3

import "github.com/bokwoon95/sq"

type CreateTableCommand struct {
	Dialect       string
	DoIfNotExists bool
	Table         Table
	Query         sq.Query
}
