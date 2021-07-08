package ddl3

type IndexDiff struct {
	TableSchema   string
	TableName     string
	IndexName     string
	IndexType     string
	CreateCommand CreateIndexCommand
	DropCommand   DropIndexCommand
	RenameCommand Command
}

type CreateIndexCommand struct {
	Valid          bool
	DoConcurrently bool
	DoIfNotExists  bool
	Index          Index
}

type DropIndexCommand struct {
	Valid          bool
	DoConcurrently bool
	DoIfExists     bool
	TableSchemas   []string
	TableNames     []string
	IndexSchemas   []string
	IndexNames     []string
	DoCascade      bool
}

type RenameIndexCommand struct {
	Valid bool
}
