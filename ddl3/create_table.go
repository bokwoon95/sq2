package ddl3

import (
	"bytes"

	"github.com/bokwoon95/sq"
)

type CreateTableCommand struct {
	Dialect       string
	DoIfNotExists bool
	Table         Table
	Query         sq.Query
}

var _ Command = CreateTableCommand{}

func (cmd CreateTableCommand) GetType() string { return CREATE_TABLE }

func (cmd CreateTableCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return nil
}
