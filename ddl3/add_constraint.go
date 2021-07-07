package ddl3

type AddConstraintCommand struct {
	Valid              bool
	AlterTableIfExists bool
	TableSchema        string
	TableName          string
	Constraint         Constraint
	DoIfNotExists      bool
	IndexSchema        string
	IndexName          string
	IsNotValid         bool
}
