package ddl3

type AlterColumnCommand struct {
	Dialect         string
	DoIfTableExists bool
	AlterType       int // TYPE | SET DEFAULT | DROP DEFAULT | NOT NULL | DROP NULL | DROP EXPRESSION | SET IDENTITY | DROP IDENTITY
	Column          Column
	Using           string
	DoIfExists      bool
}

// When analyzing the Column fields, we need a magic string that represents
// "DROP this" for something like ColumnDefault. Basically an empty string
// means "ignore this field", while the magic string means "actually drop this
// field". But what magic value to use? NULL "\x00" will not serialize well to
// JSON. If we are using magic strings, we will not need the AlterType field
// anymore.
