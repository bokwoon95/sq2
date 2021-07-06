package ddl3

type CreateViewCommand struct {
	Dialect       string
	DoOrReplace   bool
	DoIfNotExists bool
	View          View
}
