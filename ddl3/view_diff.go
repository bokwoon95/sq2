package ddl3

type ViewDiff struct {
	ViewSchema     string
	ViewName       string
	CreateCommand  *CreateViewCommand
	DropCommand    *DropViewCommand
	RenameCommand  *RenameViewCommand
	ReplaceCommand *RenameViewCommand
	ViewTriggers   []TriggerDiff
}

type CreateViewCommand struct {
	View View
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
