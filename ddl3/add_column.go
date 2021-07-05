package ddl3

type AddColumnCommand struct {
	Dialect         string
	DoIfTableExists bool
	TableSchema     string
	TableName       string
	DoIfNotExists   bool
	Column          Column
	Constraints     []Constraint
}
