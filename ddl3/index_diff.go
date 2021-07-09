package ddl3

type IndexDiff struct {
	TableSchema    string
	TableName      string
	IndexName      string
	IndexType      string
	CreateCommand  *CreateIndexCommand
	DropCommand    *DropIndexCommand
	RenameCommand  *RenameIndexCommand
	ReplaceCommand *RenameIndexCommand
}

type CreateIndexCommand struct {
	CreateConcurrently bool
	CreateIfNotExists  bool
	Index              Index
}

type DropIndexCommand struct {
	DropConcurrently bool
	DropIfExists     bool
	TableSchema      string
	TableName        string
	IndexName        string
	DropCascade      bool
}

type RenameIndexCommand struct {
	AlterIndexIfExists bool
	TableSchema        string
	TableName          string
	IndexName          string
	RenameToName       string
}
