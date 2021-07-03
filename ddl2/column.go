package ddl2

type Column struct {
	TableSchema              string
	TableName                string
	ColumnName               string
	ColumnType               string
	NormalizedColumnType     string
	Precision                int
	Scale                    int
	Identity                 string
	Autoincrement            bool
	IsNotNull                bool
	IsUnique                 bool
	IsPrimaryKey             bool
	OnUpdateCurrentTimestamp bool
	GeneratedExpr            string
	GeneratedExprStored      bool
	CollationName            string
	ColumnDefault            string
	Ignore                   bool
}
