package sq

import "bytes"

type TableInfo struct {
	TableSchema string
	TableName   string
	TableAlias  string
	_           struct{}
}

func NewTableInfo(tableSchema, tableName, tableAlias string) TableInfo {
	return TableInfo{
		TableSchema: tableSchema,
		TableName:   tableName,
		TableAlias:  tableAlias,
	}
}

var _ Table = TableInfo{}

func (tbl TableInfo) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	if tbl.TableSchema != "" {
		buf.WriteString(QuoteIdentifier(dialect, tbl.TableSchema) + ".")
	}
	buf.WriteString(QuoteIdentifier(dialect, tbl.TableName))
	return nil
}

func (tbl TableInfo) GetAlias() string { return tbl.TableAlias }

func (tbl TableInfo) GetName() string { return tbl.TableName }

func (tbl TableInfo) GetSchema() string { return tbl.TableSchema }

func (tbl TableInfo) GetTableInfo() (TableInfo, error) { return tbl, nil }

type customTable struct {
	format string
	values []interface{}
}

var _ Table = customTable{}

func Tablef(format string, values ...interface{}) customTable {
	return customTable{format: format, values: values}
}

func (tbl customTable) GetAlias() string { return "" }

func (tbl customTable) GetName() string { return "" }

func (tbl customTable) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	return BufferPrintf(dialect, buf, args, params, nil, tbl.format, tbl.values)
}
