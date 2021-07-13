package ddl

import "bytes"

type TriggerMigration struct {
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

func (cmd *CreateTriggerCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(cmd.Trigger.SQL)
	return nil
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
