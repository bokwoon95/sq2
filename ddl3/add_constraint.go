package ddl3

type AddConstraintCommand struct {
	Valid           bool
	Dialect         string
	DoIfTableExists bool
	TableSchema     string
	TableName       string
	Constraint      Constraint
	DoIfNotExists   bool
	IndexSchema     string
	IndexName       string
	IsNotValid      bool
}
