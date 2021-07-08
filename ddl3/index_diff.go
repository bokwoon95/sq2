package ddl3

type IndexDiff struct {
	TableSchema   string
	TableName     string
	IndexName     string
	IndexType     string
	CreateCommand CreateIndexCommand
	DropCommand   DropIndexCommand
	RenameCommand RenameIndexCommand
}

type CreateIndexCommand struct {
	Valid              bool
	CreateConcurrently bool
	CreateIfNotExists  bool
	Index              Index
}

type DropIndexCommand struct {
	Valid            bool
	DropConcurrently bool
	DropIfExists     bool
	TableSchema      string
	TableName        string
	IndexName        string
	DropCascade      bool
}

type RenameIndexCommand struct {
	Valid              bool
	AlterIndexIfExists bool
	TableSchema        string
	TableName          string
	IndexName          string
	RenameToName       string
}
