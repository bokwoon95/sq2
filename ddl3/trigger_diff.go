package ddl3

type TriggerDiff struct {
	TableSchema   string
	TableName     string
	TriggerName   string
	CreateCommand CreateTriggerCommand
	DropCommand   DropTriggerCommand
	RenameCommand RenameTriggerCommand
}

type CreateTriggerCommand struct {
	Valid   bool
	Trigger Trigger
}

type DropTriggerCommand struct {
	Valid        bool
	DropIfExists bool
	TableSchema  string
	TableName    string
	TriggerName  string
	DropCascade  bool
}

type RenameTriggerCommand struct {
	Valid        bool
	TableSchema  string
	TableName    string
	RenameToName string
}
