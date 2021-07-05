package ddl3

type AlterColumnCommand struct {
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
