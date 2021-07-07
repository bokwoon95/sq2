package ddl3

type AlterColumnCommand struct {
	Valid           bool
	DoIfTableExists bool
	DropDefault     bool
	DropNull        bool
	DropExpr        bool
	DropIdentity    bool
	Column          Column
	Using           string
	DoIfExists      bool
}
