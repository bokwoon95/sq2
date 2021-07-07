package ddl3

type AlterColumnCommand struct {
	Valid           bool
	Dialect         string
	DoIfTableExists bool
	DropDefault     bool
	DropNull        bool
	DropExpr        bool
	DropIdentity    bool
	Column          Column
	Using           string
	DoIfExists      bool
}
