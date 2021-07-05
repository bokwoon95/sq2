package ddl3

type CreateViewCommand struct {
	Dialect        string
	IsMaterialized bool
	DoOrReplace    bool
	DoIfNotExists  bool
	IsRecursive    bool
	View           View
}
