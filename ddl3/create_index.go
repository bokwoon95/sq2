package ddl3

type CreateIndexCommand struct {
	Valid          bool
	Dialect        string
	DoConcurrently bool
	DoIfNotExists  bool
	Index          Index
}
