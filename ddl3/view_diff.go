package ddl3

import "bytes"

type ViewDiff struct {
	ViewSchema     string
	ViewName       string
	CreateCommand  *CreateViewCommand
	DropCommand    *DropViewCommand
	RenameCommand  *RenameViewCommand
	ReplaceCommand *RenameViewCommand
	TriggerDiffs   []TriggerDiff
}

type CreateViewCommand struct {
	View View
}

func (cmd *CreateViewCommand) AppendSQL(dialect string, buf *bytes.Buffer, args *[]interface{}, params map[string][]int) error {
	buf.WriteString(cmd.View.SQL)
	return nil
}

type DropViewCommand struct {
	DropIfExists bool
	ViewSchemas  []string
	ViewNames    []string
	DropCascade  bool
}

type RenameViewCommand struct {
	AlterViewIfExists bool
	ViewSchema        string
	ViewName          string
	RenameToName      string
}
