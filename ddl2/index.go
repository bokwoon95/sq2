package ddl2

type Index struct {
	IndexSchema string
	IndexName   string
	IndexType   string
	IsUnique    bool
	TableSchema string
	TableName   string
	Columns     []string
	Exprs       []string
	Include     []string
	Where       string
}
