package ddl3

type ViewDiff struct {
	ViewSchema    string
	ViewName      string
	CreateCommand CreateViewCommand
	DropCommand   DropViewCommand
	RenameCommand RenameViewCommand
}

type CreateViewCommand struct {
	Valid bool
	View  View
}

type DropViewCommand struct {
	Valid        bool
	DropIfExists bool
	ViewSchemas  []string
	ViewNames    []string
	DoCascade    bool
}

type RenameViewCommand struct {
	Valid         bool
	AlterIfExists bool
	ViewSchema    string
	ViewName      string
	RenameToName  string
}
