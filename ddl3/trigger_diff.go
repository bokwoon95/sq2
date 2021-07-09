package ddl3

type TriggerDiff struct {
	TableSchema   string
	TableName     string
	TriggerName   string
	CreateCommand *CreateTriggerCommand
	DropCommand   *DropTriggerCommand
	RenameCommand *RenameTriggerCommand
}

type CreateTriggerCommand struct {
	Trigger Trigger
}

type DropTriggerCommand struct {
	DropIfExists bool
	TableSchema  string
	TableName    string
	TriggerName  string
	DropCascade  bool
}

type RenameTriggerCommand struct {
	TableSchema  string
	TableName    string
	RenameToName string
}
