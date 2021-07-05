package ddl3

type CreateIndexCommand struct {
	Dialect        string
	DoConcurrently bool
	DoIfNotExists  bool
	Index          Index
}
