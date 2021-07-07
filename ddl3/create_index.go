package ddl3

type CreateIndexCommand struct {
	Valid          bool
	DoConcurrently bool
	DoIfNotExists  bool
	Index          Index
}
