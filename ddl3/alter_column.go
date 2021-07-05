package ddl3

type AlterColumnCommand struct {
	Dialect         string
	DoIfTableExists bool
	AlterType       string // TYPE | SET DEFAULT | DROP DEFAULT | NOT NULL | DROP NULL | DROP EXPRESSION | SET IDENTITY | DROP IDENTITY
	OldColumnName   string
	NewColumn       Column
	DoIfExists      bool
}
