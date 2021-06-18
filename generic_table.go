package sq

import "bytes"

type GenericTable struct {
	TableSchema string
	TableName   string
	TableAlias  string
	_           struct{}
}

func NewGenericTable(tableSchema, tableName, tableAlias string) GenericTable {
	return GenericTable{
		TableSchema: tableSchema,
		TableName:   tableName,
		TableAlias:  tableAlias,
	}
}

var _ SQLAppender = GenericTable{}

func (tbl GenericTable) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if tbl.TableSchema != "" {
		buf.WriteString(QuoteIdentifier(dialect, tbl.TableSchema) + ".")
	}
	buf.WriteString(QuoteIdentifier(dialect, tbl.TableName))
	return nil
}

func (tbl GenericTable) GetAlias() string { return tbl.TableAlias }

func (tbl GenericTable) GetName() string { return tbl.TableName }

func (tbl GenericTable) GetSchema() string { return tbl.TableSchema }
