package ddl3

type CreateViewCommand struct {
	Valid   bool
	Dialect string
	View    View
}
